package game

import (
	"sync"

	"github.com/google/uuid"
)

type GameStore struct {
	mu sync.RWMutex
	ID map[string]*Game
}

var Store = &GameStore{
	ID: make(map[string]*Game),
}

func GenerateSessionID() string {
	return uuid.NewString()

}

func (s *GameStore) GetGame(sessionID string) *Game {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.ID[sessionID]
	if !exists {
		game = NewGame()
		s.ID[sessionID] = game
	}

	return game
}
