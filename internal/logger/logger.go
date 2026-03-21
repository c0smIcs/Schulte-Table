package logger

import (
	"log/slog"
	"os"
)

func InitLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				return slog.String(slog.TimeKey, t.Format("2006-01-02 15:04:05"))
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stderr, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func LoggerDBError(err error, dsnM string) {
	slog.Error("не удалось подключиться к база данных",
		slog.String("error", err.Error()),
		slog.Group("config",
			slog.String("DSN", dsnM),
		),
	)
}

func LoggerDBConnect(dsn string) {
	slog.Info("БД успешно подключена",
		slog.Group("config",
			slog.String("DSN", dsn),
		),
	)
}

func WithSession(sessionID string) *slog.Logger {
	return slog.With("session_id", sessionID)
}
