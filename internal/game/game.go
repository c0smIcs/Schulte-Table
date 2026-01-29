package game

import (
	"math/rand/v2"
	"time"
)

type Game struct {
	Board      [][]int
	NextNumber int
	Status     string
	StartTime  time.Time
}

func NewGame() *Game {
	matrix := GenerateBoard()

	return &Game{
		Board:      matrix,
		NextNumber: 1,
		Status:     "Playing",
		StartTime:  time.Now(),
	}
}

func GenerateBoard() [][]int {
	nums := rand.Perm(25)
	for i := range nums {
		nums[i]++
	}

	matrix := make([][]int, 5)
	for i := 0; i < 5; i++ {
		matrix[i] = nums[i*5 : (i+1)*5]
	}
	return matrix

}
