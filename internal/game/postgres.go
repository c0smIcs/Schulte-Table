package game

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/c0smIcs/SchulteTable/internal/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*

для запуска контейнера
docker run --name schulte-table -e POSTGRES_PASSWORD='1212' -p 5433:5432 -d --rm postgres

применение миграции к БД:
migrate -path ./migrations -database 'postgres://postgres:1212@localhost:5433/postgres?sslmode=disable' up

up - означает запустить все файлы миграции с постфиксом up

*/

type GameStore struct {
	mu   sync.RWMutex
	ID   map[string]*Game
	pool *pgxpool.Pool
}

var Store = &GameStore{
	ID: make(map[string]*Game),
}

func InitDB(dsn string) (*pgxpool.Pool, error) {
	// 1. создание конфига
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 25
	config.MinConns = 1
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 3 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second

	// 2. создаем пул через pgxpool.NewWithConfig
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		logger.LoggerDBError(err, config.ConnString())
		return nil, err
	}

	// 3. делаем Ping, чтобы проверить связь
	err = conn.Ping(ctx)
	if err != nil {
		return nil, err
	}

	Store.pool = conn

	logger.LoggerDBConnect(config.ConnString())

	return conn, nil
}

func GenerateSessionID() string {
	return uuid.NewString()
}

func (s *GameStore) GetGame(sessionID string) *Game {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	return s.ID[sessionID]
}

func (gs *GameStore) SaveGame(sessionID string, game *Game) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.ID[sessionID] = game
}

func (gs *GameStore) SaveRecord(ctx context.Context, sessionID string, duration time.Duration) error {
	query := `INSERT INTO record(SessionID, TimeTaken) VALUES ($1, $2)`
	
	_, err := gs.pool.Exec(ctx, query, sessionID, duration.Seconds())
	if err != nil {
		return err
	}

	return nil
}

func (gs *GameStore) GetBestTime(ctx context.Context, sessionID string) (string, error) {
	query := `SELECT TimeTaken FROM record WHERE SessionID = $1 ORDER BY TimeTaken ASC LIMIT 1`

	var bestTime float64

	err := gs.pool.QueryRow(ctx, query, sessionID).Scan(&bestTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "--:--", nil
		}
		return "", err
	}

	duration := time.Duration(bestTime * float64(time.Second))

	return FormatDuration(duration), nil
}
