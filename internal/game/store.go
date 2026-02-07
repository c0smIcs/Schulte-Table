package game

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)
type GameStore struct {
	mu sync.RWMutex
	ID map[string]*Game
}

type Record struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	SessionID string    `gorm:"index"`
	TimeTaken float64
	CreatedAt time.Time
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

// хук GORM для генерации UUID перед созданием
func (r *Record) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID, err = uuid.NewRandom()
	return
}

func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	// Автомиграция создает таблицу автоматически
	if err := db.AutoMigrate(&Record{}); err != nil {
		return nil, fmt.Errorf("migration failed %w", err)
	}

	return db, nil
}

func SaveRecord(ctx context.Context, db *gorm.DB, sessionID string, duration time.Duration) error {
	record := &Record{
		SessionID: sessionID,
		TimeTaken: duration.Seconds(),
	}

	result := db.WithContext(ctx).Create(record)
	return result.Error
}

func GetBestTime(ctx context.Context, db *gorm.DB, sessionID string) (string, error) {
	var rec Record
	// Ищем запись с минимальным TimeTaken для этого sessionID
	err := db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("time_taken ASC").
		First(&rec).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "--:--", nil
		}
		return "", nil
	}

	d := time.Duration(rec.TimeTaken * float64(time.Second))
	return FormatDuration(d), err
}

