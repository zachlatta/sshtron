package main

import (
	"fmt"
	"io"
)

type Hub struct {
	Sessions   map[*Session]struct{}
	Rerender   chan struct{}
	Register   chan *Session
	Unregister chan *Session
}

func NewHub() Hub {
	return Hub{
		Sessions:   make(map[*Session]struct{}),
		Rerender:   make(chan struct{}),
		Register:   make(chan *Session),
		Unregister: make(chan *Session),
	}
}

func (h *Hub) Run(g *Game) {
	for {
		select {
		case <-h.Rerender:
			for s := range h.Sessions {
				g.Render(s)
			}
		case s := <-h.Register:
			h.Sessions[s] = struct{}{}
		case s := <-h.Unregister:
			if _, ok := h.Sessions[s]; ok {
				delete(h.Sessions, s)
			}
		}
	}
}

type Position struct {
	X int
	Y int
}

type Player struct {
	Marker rune
	Pos    *Position
}

func NewPlayer() *Player {
	return &Player{Marker: 'x', Pos: &Position{0, 0}}
}

type TileType int

const (
	TileGrass TileType = iota
	TileBlocker
)

type Tile struct {
	Type TileType
}

type World struct {
	// Top left is 0,0
	level [][]Tile
}

type Game struct {
	hub Hub

	// Top left is 0,0
	level [][]Tile
}

func NewGame(worldWidth, worldHeight int) *Game {
	g := &Game{hub: NewHub()}
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
			g.setTileType(Position{x, y}, TileGrass)
		}
	}
}

func (g *Game) setTileType(pos Position, tileType TileType) error {
	outOfBoundsErr := "The given %s value (%s) is out of bounds"
	if pos.X > len(g.level) || pos.X < 0 {
		return fmt.Errorf(outOfBoundsErr, "X", pos.X)
	} else if pos.Y > len(g.level[pos.X]) || pos.Y < 0 {
		return fmt.Errorf(outOfBoundsErr, "Y", pos.Y)
	}

	g.level[pos.X][pos.Y].Type = tileType

	return nil
}

func (g *Game) players() map[*Player]struct{} {
	players := make(map[*Player]struct{})

	for session := range g.hub.Sessions {
		players[session.Player] = struct{}{}
	}

	return players
}

// Warning: this will only work with square worlds
func (g *Game) worldString() string {
	str := ""
	worldWidth := len(g.level)
	worldHeight := len(g.level[0])

	// Create two dimensional slice of runes to represent the world
	strWorld := make([][]rune, worldWidth)
	for x := range strWorld {
		strWorld[x] = make([]rune, worldHeight)
	}

	// Load the level into the rune slice
	for x := 0; x < worldWidth; x++ {
		for y := 0; y < worldHeight; y++ {
			tile := g.level[x][y]

			switch tile.Type {
			case TileGrass:
				strWorld[x][y] = '□'
			case TileBlocker:
				strWorld[x][y] = '■'
			}
		}
	}

	// Load the players into the rune slice
	for player := range g.players() {
		pos := player.Pos
		strWorld[pos.X][pos.Y] = 'x'
	}

	// Convert the rune slice to a string
	for y := 0; y < worldHeight; y++ {
		for x := 0; x < worldWidth; x++ {
			str += string(strWorld[x][y])
		}

		str += "\r\n"
	}

	return str
}

func (g *Game) Run() {
	g.hub.Run(g)
}

func (g *Game) Render(w io.Writer) {
	worldStr := g.worldString()

	fmt.Fprint(w, worldStr)
}

func (g *Game) AddSession(s *Session) {
	g.hub.Register <- s
}

type Session struct {
	c io.ReadWriter

	Player *Player
}

func NewSession(c io.ReadWriter) *Session {
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
