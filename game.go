package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"time"
)

type Hub struct {
	Sessions   map[*Session]struct{}
	Redraw     chan struct{}
	Register   chan *Session
	Unregister chan *Session
}

func NewHub() Hub {
	return Hub{
		Sessions:   make(map[*Session]struct{}),
		Redraw:     make(chan struct{}),
		Register:   make(chan *Session),
		Unregister: make(chan *Session),
	}
}

func (h *Hub) Run(g *Game) {
	for {
		select {
		case <-h.Redraw:
			for s := range h.Sessions {
				g.Render(s)
			}
		case s := <-h.Register:
			h.Sessions[s] = struct{}{}
		case s := <-h.Unregister:
			if _, ok := h.Sessions[s]; ok {
				fmt.Fprint(s, "End of line.\r\n\r\n")
				delete(h.Sessions, s)
				s.c.Close()
			}
		}
	}
}

type Position struct {
	X float64
	Y float64
}

func (p Position) RoundX() int {
	return int(p.X + 0.5)
}

func (p Position) RoundY() int {
	return int(p.Y + 0.5)
}

type PlayerDirection int

const (
	verticalPlayerSpeed   = 0.007
	horizontalPlayerSpeed = 0.01

	playerUpRune    = '⇧'
	playerLeftRune  = '⇦'
	playerDownRune  = '⇩'
	playerRightRune = '⇨'

	PlayerUp PlayerDirection = iota
	PlayerLeft
	PlayerDown
	PlayerRight
)

type Player struct {
	Direction PlayerDirection
	Marker    rune
	Pos       *Position
}

func NewPlayer() *Player {
	return &Player{
		Marker:    playerDownRune,
		Direction: PlayerDown,
		Pos:       &Position{0, 0},
	}
}

func (p *Player) HandleUp() {
	p.Direction = PlayerUp
	p.Marker = playerUpRune
}

func (p *Player) HandleLeft() {
	p.Direction = PlayerLeft
	p.Marker = playerLeftRune
}

func (p *Player) HandleDown() {
	p.Direction = PlayerDown
	p.Marker = playerDownRune
}

func (p *Player) HandleRight() {
	p.Direction = PlayerRight
	p.Marker = playerRightRune
}

func (p *Player) Update(delta float64) {
	switch p.Direction {
	case PlayerUp:
		p.Pos.Y -= verticalPlayerSpeed * delta
	case PlayerLeft:
		p.Pos.X -= horizontalPlayerSpeed * delta
	case PlayerDown:
		p.Pos.Y += verticalPlayerSpeed * delta
	case PlayerRight:
		p.Pos.X += horizontalPlayerSpeed * delta
	}
}

type TileType int

const (
	TileGrass TileType = iota
	TileBlocker
)

type Tile struct {
	Type TileType
}

type Game struct {
	hub Hub

	Redraw chan struct{}

	// Top left is 0,0
	level [][]Tile
}

func NewGame(worldWidth, worldHeight int) *Game {
	g := &Game{
		hub:    NewHub(),
		Redraw: make(chan struct{}),
	}
	g.initalizeLevel(worldWidth, worldHeight)

	return g
}

func (g *Game) initalizeLevel(width, height int) {
	g.level = make([][]Tile, width)
	for x := range g.level {
		g.level[x] = make([]Tile, height)
	}

	// Default world to grass
	for x := range g.level {
		for y := range g.level[x] {
			g.setTileType(Position{float64(x), float64(y)}, TileGrass)
		}
	}
}

func (g *Game) setTileType(pos Position, tileType TileType) error {
	outOfBoundsErr := "The given %s value (%s) is out of bounds"
	if pos.RoundX() > len(g.level) || pos.RoundX() < 0 {
		return fmt.Errorf(outOfBoundsErr, "X", pos.X)
	} else if pos.RoundY() > len(g.level[pos.RoundX()]) || pos.RoundY() < 0 {
		return fmt.Errorf(outOfBoundsErr, "Y", pos.Y)
	}

	g.level[pos.RoundX()][pos.RoundY()].Type = tileType

	return nil
}

func (g *Game) players() map[*Player]*Session {
	players := make(map[*Player]*Session)

	for session := range g.hub.Sessions {
		players[session.Player] = session
	}

	return players
}

// Characters for rendering
const (
	verticalWall   = '┆'
	horizontalWall = '┄'
	topLeft        = '╭'
	topRight       = '╮'
	bottomRight    = '╯'
	bottomLeft     = '╰'

	grass   = ' '
	blocker = '■'
)

// Warning: this will only work with square worlds
func (g *Game) worldString() string {
	str := ""
	worldWidth := len(g.level)
	worldHeight := len(g.level[0])

	// Create two dimensional slice of runes to represent the world. It's two
	// characters larger in each direction to accomodate for walls.
	strWorld := make([][]rune, worldWidth+3)
	for x := range strWorld {
		strWorld[x] = make([]rune, worldHeight+3)
	}

	// Load the walls into the rune slice
	for x := 0; x < worldWidth+2; x++ {
		strWorld[x][0] = horizontalWall
		strWorld[x][worldHeight+1] = horizontalWall
	}
	for y := 0; y < worldHeight+2; y++ {
		strWorld[0][y] = verticalWall
		strWorld[worldWidth+1][y] = verticalWall
	}

	// Time for the edges!
	strWorld[0][0] = topLeft
	strWorld[worldWidth+1][0] = topRight
	strWorld[worldWidth+1][worldHeight+1] = bottomRight
	strWorld[0][worldHeight+1] = bottomLeft

	// Load the level into the rune slice
	for x := 0; x < worldWidth; x++ {
		for y := 0; y < worldHeight; y++ {
			tile := g.level[x][y]

			switch tile.Type {
			case TileGrass:
				strWorld[x+1][y+1] = grass
			case TileBlocker:
				strWorld[x+1][y+1] = blocker
			}
		}
	}

	// Load the players into the rune slice
	for player := range g.players() {
		pos := player.Pos
		strWorld[pos.RoundX()+1][pos.RoundY()+1] = player.Marker
	}

	// Convert the rune slice to a string
	for y := 0; y < len(strWorld[0]); y++ {
		for x := 0; x < len(strWorld); x++ {
			str += string(strWorld[x][y])
		}

		str += "\r\n"
	}

	return str
}

func (g *Game) Run() {
	// Proxy g.Redraw's channel to g.hub.Redraw
	go func() {
		for {
			g.hub.Redraw <- <-g.Redraw
		}
	}()

	// Run game loop
	go func() {
		var lastUpdate time.Time

		c := time.Tick(time.Second / 60)
		for now := range c {
			g.Update(float64(now.Sub(lastUpdate)) / float64(time.Millisecond))

			lastUpdate = now
		}
	}()

	// Redraw regularly.
	//
	// TODO: Implement diffing and only redraw when needed
	go func() {
		c := time.Tick(time.Second / 10)
		for range c {
			g.Redraw <- struct{}{}
		}
	}()

	g.hub.Run(g)
}

// Update is the main game logic loop. Delta is the time since the last update
// in milliseconds.
func (g *Game) Update(delta float64) {
	for player, session := range g.players() {
		player.Update(delta)

		// Kick player if they're out of bounds
		pos := player.Pos
		if pos.RoundX() < 0 || pos.RoundX() > len(g.level) ||
			pos.RoundY() < 0 || pos.RoundY() > len(g.level[0]) {
			g.hub.Unregister <- session
		}
	}
}

func (g *Game) Render(w io.Writer) {
	worldStr := g.worldString()

	fmt.Fprintln(w)
	fmt.Fprint(w, worldStr)
}

func (g *Game) AddSession(s *Session) {
	g.hub.Register <- s
}

type Session struct {
	c ssh.Channel

	Player *Player
}

func NewSession(c ssh.Channel) *Session {
	s := Session{c: c}
	s.Player = NewPlayer()

	return &s
}

func (s *Session) Read(p []byte) (int, error) {
	return s.c.Read(p)
}

func (s *Session) Write(p []byte) (int, error) {
	return s.c.Write(p)
}
