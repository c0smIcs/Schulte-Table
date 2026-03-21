package game

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

type Game struct {
	Board      [][]int
	NextNumber int
	Status     string
	StartTime  time.Time
	RWmu       sync.RWMutex
}

func NewGame(sessionID string) *Game {
	matrix := GenerateBoard()

	game := &Game{
		Board:      matrix,
		NextNumber: 1,
		Status:     "Playing",
		StartTime:  time.Now(),
	}

	return game
}

func GenerateBoard() [][]int {
	generateNums := rand.Perm(25)
	for i := range generateNums {
		generateNums[i]++
	}

	matrix := make([][]int, 5)
	for i := 0; i < 5; i++ {
		matrix[i] = generateNums[i*5 : (i+1)*5]
	}

	return matrix
}

func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = d.Abs()
	}

	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := d.Milliseconds() % 1000

	timeStr := fmt.Sprintf("%02d:%02d:%03d", minutes, seconds, milliseconds)

	return timeStr
}

func (g *Game) Reset() {
	matrix := GenerateBoard()

	g.RWmu.Lock()
	defer g.RWmu.Unlock()
	g.NextNumber = 1
	g.Status = "Playing"
	g.StartTime = time.Now()
	g.Board = matrix
}
