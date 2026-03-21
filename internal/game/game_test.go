package game

import (
	"testing"
	"time"
)

func TestNewGame(t *testing.T) {
	game := NewGame("test-session-123")
	if game == nil {
		t.Fatal("NewGame() вернул nil, ожидался *Game")
	}

	if game.NextNumber != 1 {
		t.Errorf("NextNumber = %d, ожидалось 1", game.NextNumber)
	}

	if game.Status != "Playing" {
		t.Errorf("Status = %q, ожидалось %q", game.Status, "Playing")
	}

	if game.StartTime.IsZero() {
		t.Error("StartTime не установлено")
	}

	if game.Board == nil {
		t.Error("Board = nil")
	}
}

func checkBoardValid(board [][]int, t *testing.T) {
	t.Helper()

	seen := make(map[int]bool)

	for i, row := range board {
		// 1. Проверяем количество колонок в каждой строке
		if len(row) != 5 {
			t.Errorf("в строке %d ожидалось 5 колонок, получили %d", i, len(row))
		}

		for _, num := range row {
			// 3. проверяем диапазон чисел от 1 до 25
			if num < 1 || num > 25 {
				t.Errorf("число %d вне диапазона [1, 25]", num)
			}

			// 4. проверяем уникальность
			if seen[num] {
				t.Errorf("число %d дублируется в матрице", num)
			}

			seen[num] = true
		}
	}

	// 5. проверяем, что всего набралось 25 уникальных чисел
	if len(seen) != 25 {
		t.Errorf("ожидалось 25 уникальных чисел, получили %d", len(seen))
	}
}

func TestGenerateBoard(t *testing.T) {
	board := GenerateBoard()

	checkBoardValid(board, t)
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{"test 1", 1*time.Minute + 2*time.Second + 300*time.Millisecond, "01:02:300"},
		{"test 2", 4*time.Minute + 5*time.Second + 500*time.Millisecond, "04:05:500"},
		{"test 3", 0*time.Minute + 0*time.Second + 0*time.Millisecond, "00:00:000"},
		{"test 4", 61*time.Minute + 10*time.Second + 205*time.Millisecond, "01:10:205"},
		{"test 5", -10*time.Minute + -10*time.Second + -100*time.Millisecond, "10:10:100"},
		{"test 6", 1*time.Minute + 1*time.Second + 999*time.Millisecond, "01:01:999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.input)
			if result != tt.expected {
				t.Errorf("FormatDuration(%v) = %s; want = %s", tt.input, result, tt.expected)
			}
		})
	}
}

/*
	 func (g *Game) Reset() {
		matrix := GenerateBoard()

		g.RWmu.Lock()
		defer g.RWmu.Unlock()
		g.NextNumber = 1
		g.Status = "Playing"
		g.StartTime = time.Now()
		g.Board = matrix
	}

Проверить, что после вызова счетчик сбрасывается в 1, а время начала обновляется.
*/
func TestReset(t *testing.T) {
	game := NewGame("test-session-id")

	game.Reset()

	if game.NextNumber != 1 {
		t.Errorf("NextNumber = %d, ожидалось 1", game.NextNumber)
	}

	if game.Status != "Playing" {
		t.Errorf("Status = %v, ожидалось 'Playing'", game.Status)
	}

	if time.Since(game.StartTime) > 1300*time.Millisecond {
		t.Errorf("StartTime = %v, ожидалось обновить время", game.StartTime)
	}

	checkBoardValid(game.Board, t)
}
