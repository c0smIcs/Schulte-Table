package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Recorder interface {
	SaveRecord(ctx context.Context, db *gorm.DB, sessionID string, duration time.Duration) error
	GetBestTime(ctx context.Context, db *gorm.DB, sessionID string) (string, error)
}

type Record struct {
	ID        uuid.UUID
	SessionID string
	TimeTaken float64
	CreatedAt time.Time
}

func (r *Record) BeforeCreate() (err error) {
	r.ID, err = uuid.NewRandom()
	if err != nil {
		return err
	}

	return
}
