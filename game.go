package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"time"

	"github.com/dustinkirkland/golang-petname"
	"github.com/fatih/color"
)

type Position struct {
	X float64
	Y float64
}

func PositionFromInt(x, y int) Position {
	return Position{float64(x), float64(y)}
}

func (p Position) RoundX() int {
	return int(p.X + 0.5)
}

func (p Position) RoundY() int {
	return int(p.Y + 0.5)
}

type PlayerDirection int

const (
	verticalPlayerSpeed        = 0.007
	horizontalPlayerSpeed      = 0.01
	playerCountScoreMultiplier = 1.25
	playerTimeout              = 15 * time.Second

	playerUpRune    = '⇡'
	playerLeftRune  = '⇠'
	playerDownRune  = '⇣'
	playerRightRune = '⇢'

	playerTrailHorizontal      = '┄'
	playerTrailVertical        = '┆'
	playerTrailLeftCornerUp    = '╭'
	playerTrailLeftCornerDown  = '╰'
	playerTrailRightCornerDown = '╯'
	playerTrailRightCornerUp   = '╮'

	playerRed     = color.FgRed
	playerGreen   = color.FgGreen
	playerMagenta = color.FgMagenta
	playerCyan    = color.FgCyan

	PlayerUp PlayerDirection = iota
	PlayerLeft
	PlayerDown
	PlayerRight
)

var playerColors = []color.Attribute{playerRed, playerGreen, playerMagenta,
	playerCyan}

var playerBorderColors = map[color.Attribute]color.Attribute{
	playerRed:     color.FgHiRed,
	playerGreen:   color.FgHiGreen,
	playerMagenta: color.FgHiMagenta,
	playerCyan:    color.FgHiCyan,
}

var playerColorNames = map[color.Attribute]string{
	playerRed:     "Red",
	playerGreen:   "Green",
	playerMagenta: "Magenta",
	playerCyan:    "Cyan",
}

type PlayerTrailSegment struct {
	Marker rune
	Pos    Position
}

type Player struct {
	s *Session

	CreatedAt  time.Time
	Direction  PlayerDirection
	Marker     rune
	LastAction time.Time
	Color      color.Attribute
	Pos        *Position

	Trail []PlayerTrailSegment

	score     float64
	HighScore int
}

// NewPlayer creates a new player. If color is below 1, a random color is chosen
func NewPlayer(s *Session, worldWidth, worldHeight int,
	color color.Attribute) *Player {

	rand.Seed(time.Now().UnixNano())

	startX := rand.Float64() * float64(worldWidth)
	startY := rand.Float64() * float64(worldHeight)

	if color < 0 {
		color = playerColors[rand.Intn(len(playerColors))]
	}

	return &Player{
		s:          s,
		CreatedAt:  time.Now(),
		Marker:     playerDownRune,
		LastAction: time.Now(),
		Direction:  PlayerDown,
		Color:      color,
		Pos:        &Position{startX, startY},
	}
}

func (p *Player) addTrailSegment(pos Position, marker rune) {
	segment := PlayerTrailSegment{marker, pos}
	p.Trail = append([]PlayerTrailSegment{segment}, p.Trail...)
}

func (p *Player) calculateScore(delta float64, playerCount int) float64 {
	rawIncrement := (delta * (float64(playerCount-1) * playerCountScoreMultiplier))

	// Convert millisecond increment to seconds
	actualIncrement := rawIncrement / 1000

	return p.score + actualIncrement
}

func (p *Player) HandleUp() {
	if p.Direction == PlayerDown {
		return
	}
	p.Direction = PlayerUp
	p.Marker = playerUpRune
	p.didAction()
}

func (p *Player) HandleLeft() {
	if p.Direction == PlayerRight {
		return
	}
	p.Direction = PlayerLeft
	p.Marker = playerLeftRune
	p.didAction()
}

func (p *Player) HandleDown() {
	if p.Direction == PlayerUp {
		return
	}
	p.Direction = PlayerDown
	p.Marker = playerDownRune
	p.didAction()
}

func (p *Player) HandleRight() {
	if p.Direction == PlayerLeft {
		return
	}
	p.Direction = PlayerRight
	p.Marker = playerRightRune
	p.didAction()
}

func (p *Player) didAction() {
	p.LastAction = time.Now()
}

func (p *Player) Score() int {
	return int(p.score)
}

func (p *Player) Update(g *Game, delta float64) {
	startX, startY := p.Pos.RoundX(), p.Pos.RoundY()

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

	endX, endY := p.Pos.RoundX(), p.Pos.RoundY()

	// If we moved, add a trail segment.
	if endX != startX || endY != startY {
		var lastSeg *PlayerTrailSegment
		var lastSegX, lastSegY int
		if len(p.Trail) > 0 {
			lastSeg = &p.Trail[0]
			lastSegX = lastSeg.Pos.RoundX()
			lastSegY = lastSeg.Pos.RoundY()
		}

		pos := PositionFromInt(startX, startY)

		switch {
		// Handle corners. This took an ungodly amount of time to figure out. Highly
		// recommend you don't touch.
		case lastSeg != nil &&
			(p.Direction == PlayerRight && endX > lastSegX && endY < lastSegY) ||
			(p.Direction == PlayerDown && endX < lastSegX && endY > lastSegY):
			p.addTrailSegment(pos, playerTrailLeftCornerUp)
		case lastSeg != nil &&
			(p.Direction == PlayerUp && endX > lastSegX && endY < lastSegY) ||
			(p.Direction == PlayerLeft && endX < lastSegX && endY > lastSegY):
			p.addTrailSegment(pos, playerTrailRightCornerDown)
		case lastSeg != nil &&
			(p.Direction == PlayerDown && endX > lastSegX && endY > lastSegY) ||
			(p.Direction == PlayerLeft && endX < lastSegX && endY < lastSegY):
			p.addTrailSegment(pos, playerTrailRightCornerUp)
		case lastSeg != nil &&
			(p.Direction == PlayerRight && endX > lastSegX && endY > lastSegY) ||
			(p.Direction == PlayerUp && endX < lastSegX && endY < lastSegY):
			p.addTrailSegment(pos, playerTrailLeftCornerDown)

		// Vertical and horizontal trails
		case endX == startX && endY < startY:
			p.addTrailSegment(pos, playerTrailVertical)
		case endX < startX && endY == startY:
			p.addTrailSegment(pos, playerTrailHorizontal)
		case endX == startX && endY > startY:
			p.addTrailSegment(pos, playerTrailVertical)
		case endX > startX && endY == startY:
			p.addTrailSegment(pos, playerTrailHorizontal)
		}
	}

	p.score = p.calculateScore(delta, len(g.players()))
}

type ByColor []*Player

func (slice ByColor) Len() int {
	return len(slice)
}

func (slice ByColor) Less(i, j int) bool {
	return playerColorNames[slice[i].Color] < playerColorNames[slice[j].Color]
}

func (slice ByColor) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type Game struct {
	Name      string
	Redraw    chan struct{}
	HighScore int

	width    int
	height   int
	Sessions map[*Session]struct{}
}

func NewGame(worldWidth, worldHeight int) *Game {
	return &Game{
		Name:     petname.Generate(1, ""),
		Redraw:   make(chan struct{}),
		Sessions: make(map[*Session]struct{}),
		width:    worldWidth,
		height:   worldHeight,
	}
}

func (g *Game) players() map[*Player]*Session {
	players := make(map[*Player]*Session)

	for session := range g.Sessions {
		players[session.Player] = session
	}

	return players
}

// Characters for rendering
const (
	verticalWall   = '║'
	horizontalWall = '═'
	topLeft        = '╔'
	topRight       = '╗'
	bottomRight    = '╝'
	bottomLeft     = '╚'
)

// Warning: this will only work with square worlds
func (g *Game) worldString(s *Session) string {
	worldWidth := g.WorldWidth()
	worldHeight := g.WorldHeight()

	// Create two dimensional slice of strings to represent the world. It's two
	// characters larger in each direction to accomodate for walls.
	strWorld := make([][]string, worldWidth+2)
	for x := range strWorld {
		strWorld[x] = make([]string, worldHeight+2)
	}

	// Load the walls into the rune slice
	borderColorizer := color.New(playerBorderColors[s.Player.Color]).SprintFunc()
	for x := 0; x < worldWidth+2; x++ {
		strWorld[x][0] = borderColorizer(string(horizontalWall))
		strWorld[x][worldHeight+1] = borderColorizer(string(horizontalWall))
	}
	for y := 0; y < worldHeight+2; y++ {
		strWorld[0][y] = borderColorizer(string(verticalWall))
		strWorld[worldWidth+1][y] = borderColorizer(string(verticalWall))
	}

	// Time for the edges!
	strWorld[0][0] = borderColorizer(string(topLeft))
	strWorld[worldWidth+1][0] = borderColorizer(string(topRight))
	strWorld[worldWidth+1][worldHeight+1] = borderColorizer(string(bottomRight))
	strWorld[0][worldHeight+1] = borderColorizer(string(bottomLeft))

	// Draw the player's score
	scoreStr := fmt.Sprintf(
		" Score: %d : Your High Score: %d : Game High Score: %d ",
		s.Player.Score(),
		s.Player.HighScore,
		g.HighScore,
	)
	for i, r := range scoreStr {
		strWorld[3+i][0] = borderColorizer(string(r))
	}

	// Draw the player's color
	colorStr := fmt.Sprintf(" %s ", playerColorNames[s.Player.Color])
	colorStrColorizer := color.New(s.Player.Color).SprintFunc()
	for i, r := range colorStr {
		charsRemaining := len(colorStr) - i
		strWorld[len(strWorld)-3-charsRemaining][0] = colorStrColorizer(string(r))
	}

	// Draw everyone's scores
	if len(g.players()) > 1 {
		// Sort the players by color name
		players := []*Player{}

		for player := range g.players() {
			if player == s.Player {
				continue
			}

			players = append(players, player)
		}

		sort.Sort(ByColor(players))
		startX := 3

		// Actually draw their scores
		for _, player := range players {
			colorizer := color.New(player.Color).SprintFunc()
			scoreStr := fmt.Sprintf(" %s: %d",
				playerColorNames[player.Color],
				player.Score(),
			)
			for _, r := range scoreStr {
				strWorld[startX][len(strWorld[0])-1] = colorizer(string(r))
				startX++
			}
		}

		// Add final spacing next to wall
		strWorld[startX][len(strWorld[0])-1] = " "
	} else {
		warning :=
			" Warning: Other Players Must be in This Game for You to Score! "
		for i, r := range warning {
			strWorld[3+i][len(strWorld[0])-1] = borderColorizer(string(r))
		}
	}

	// Draw the game's name
	nameStr := fmt.Sprintf(" %s ", g.Name)
	for i, r := range nameStr {
		charsRemaining := len(nameStr) - i
		strWorld[len(strWorld)-3-charsRemaining][len(strWorld[0])-1] =
			borderColorizer(string(r))
	}

	for x := 1; x <= worldWidth; x++ {
		for y := 1; y <= worldHeight; y++ {
			strWorld[x][y] = " "
		}
	}

	// Load the players into the rune slice
	for player := range g.players() {
		colorizer := color.New(player.Color).SprintFunc()

		pos := player.Pos
		strWorld[pos.RoundX()+1][pos.RoundY()+1] = colorizer(string(player.Marker))

		// Load the player's trail into the rune slice
		for _, segment := range player.Trail {
			x, y := segment.Pos.RoundX()+1, segment.Pos.RoundY()+1
			strWorld[x][y] = colorizer(string(segment.Marker))
		}
	}

	// Convert the rune slice to a string
	buffer := bytes.NewBuffer(make([]byte, 0, worldWidth*worldHeight*2))
	for y := 0; y < len(strWorld[0]); y++ {
		for x := 0; x < len(strWorld); x++ {
			buffer.WriteString(strWorld[x][y])
		}

		// Don't add an extra newline if we're on the last iteration
		if y != len(strWorld[0])-1 {
			buffer.WriteString("\r\n")
		}
	}

	return buffer.String()
}

func (g *Game) WorldWidth() int {
	return g.width
}

func (g *Game) WorldHeight() int {
	return g.height
}

func (g *Game) AvailableColors() []color.Attribute {
	usedColors := map[color.Attribute]bool{}
	for _, color := range playerColors {
		usedColors[color] = false
	}

	for player := range g.players() {
		usedColors[player.Color] = true
	}

	availableColors := []color.Attribute{}
	for color, used := range usedColors {
		if !used {
			availableColors = append(availableColors, color)
		}
	}

	return availableColors
}

func (g *Game) SessionCount() int {
	return len(g.Sessions)
}

func (g *Game) Run() {
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
			for s := range g.Sessions {
				go g.Render(s)
			}
		}
	}()
}

// Update is the main game logic loop. Delta is the time since the last update
// in milliseconds.
func (g *Game) Update(delta float64) {
	// We'll use this to make a set of all of the coordinates that are occupied by
	// trails
	trailCoordMap := make(map[string]bool)

	// Update player data
	for player, session := range g.players() {
		player.Update(g, delta)

		// Update session high score, if applicable
		if player.Score() > player.HighScore {
			player.HighScore = player.Score()
		}

		// Update global high score, if applicable
		if player.Score() > g.HighScore {
			g.HighScore = player.Score()
		}

		// Restart the player if they're out of bounds
		pos := player.Pos
		if pos.RoundX() < 0 || pos.RoundX() >= g.WorldWidth() ||
			pos.RoundY() < 0 || pos.RoundY() >= g.WorldHeight() {
			session.StartOver(g.WorldWidth(), g.WorldHeight())
		}

		// Kick the player if they've timed out
		if time.Now().Sub(player.LastAction) > playerTimeout {
			fmt.Fprint(session, "\r\n\r\nYou were terminated due to inactivity\r\n")
			g.RemoveSession(session)
			return
		}

		for _, seg := range player.Trail {
			coordStr := fmt.Sprintf("%d,%d", seg.Pos.RoundX(), seg.Pos.RoundY())
			trailCoordMap[coordStr] = true
		}
	}

	// Check if any players collide with a trail and restart them if so
	for player, session := range g.players() {
		playerPos := fmt.Sprintf("%d,%d", player.Pos.RoundX(), player.Pos.RoundY())
		if collided := trailCoordMap[playerPos]; collided {
			session.StartOver(g.WorldWidth(), g.WorldHeight())
		}
	}
}

func (g *Game) Render(s *Session) {
	worldStr := g.worldString(s)

	var b bytes.Buffer
	b.WriteString("\033[H\033[2J")
	b.WriteString(worldStr)

	// Send over the rendered world
	io.Copy(s, &b)
}

func (g *Game) AddSession(s *Session) {
	// Hide the cursor
	fmt.Fprint(s, "\033[?25l")

	g.Sessions[s] = struct{}{}
}

func (g *Game) RemoveSession(s *Session) {
	if _, ok := g.Sessions[s]; ok {
		fmt.Fprint(s, "\r\n\r\n~ End of Line ~ \r\n\r\nRemember to use WASD to move!\r\n\r\n")

		// Unhide the cursor
		fmt.Fprint(s, "\033[?25h")

		delete(g.Sessions, s)
		s.c.Close()
	}
}
