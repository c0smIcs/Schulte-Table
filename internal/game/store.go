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

// Старый код:
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

// -----------------------------------------------------------------------------------------------------------------------------------

// С использованием GORM:
/*
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	SessionID string    `gorm:"index"`
	TimeTaken float64
	CreatedAt time.Time
}

func initDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{PrepareStmt: true})
	if err != nil {
		return nil, fmt.Errorf("Ошибка соединения с БД: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		sqlDB.SetMaxIdleConns(10)           // Максимум простаивающих соединений
		sqlDB.SetMaxOpenConns(100)          // Максимум открытых соединений
		sqlDB.SetConnMaxLifetime(time.Hour) // время жизни соединения
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, fmt.Errorf("ошибка миграции БД: %w", err)
	}

	return db, nil
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	newID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	u.ID = newID

	return nil
}

func CreateUser(db *gorm.DB, record time.Time) (*User, error) {
	user := &User{
		Record: record,
	}

	ctx := context.Background()
	err := gorm.G[User](db).Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при создании: %w", err)
	}

	return user, err
}

func GetUserRecord(db *gorm.DB, record time.Time) (*User, error) {
	ctx := context.Background()

	user, err := gorm.G[*User](db).Where("record = ?", record).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	return user, nil
}

func GetUserByID(db *gorm.DB, id uuid.UUID) (*User, error) {
	ctx := context.Background()

	user, err := gorm.G[*User](db).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}

		return nil, err
	}

	return user, nil
}

func UpdateUserRecord(db *gorm.DB, id uuid.UUID, newRecord time.Time) error {
	ctx := context.Background()

	_, err := gorm.G[*User](db).Where("id = ?", id).Update(ctx, "record", newRecord)
	if err != nil {
		return fmt.Errorf("не удалось обновить новый рекорд: %w", err)
	}

	return nil
}
*/

type Record struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	SessionID string    `gorm:"index"`
	TimeTaken float64
	CreatedAt time.Time
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

