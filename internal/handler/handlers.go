package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	g "github.com/c0smIcs/SchulteTable/internal/game"
	"github.com/c0smIcs/SchulteTable/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClickResponse struct {
	IsCorrect  bool   `json:"is_correct"`
	NextNumber int    `json:"next_number"`
	Status     string `json:"status"`
	TimeTaken  string `json:"time_taken"`
	BestRecord string `json:"best_record"`
}

type App struct {
	DB *pgxpool.Pool   `json:"db"`
	WG *sync.WaitGroup `json:"wg"`
}

var tpl = template.Must(template.ParseFiles("ui/html/index.html"))

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	errorResponse := ErrorResponse{Error: msg}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		slog.Error("не удалось закодировать ошибку", "error", err)
		return
	}
}

func getGameFromRequest(w http.ResponseWriter, r *http.Request) (*g.Game, string, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		writeJSONError(w, http.StatusForbidden, "сессия не найдена")
		return nil, "", err
	}

	game := g.Store.GetGame(cookie.Value)
	if game == nil {
		writeJSONError(w, http.StatusNotFound, "не удалось найти игру")
		return nil, "", fmt.Errorf("game not found for session: %s", cookie.Value)
	}

	return game, cookie.Value, nil
}

func (a *App) IndexHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	var sessionID string

	if err != nil {
		sessionID = g.GenerateSessionID()
		slog.Info("Создана новая сессия", "session_id", sessionID)

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
		})
	} else {
		sessionID = cookie.Value
	}

	currentGame := g.Store.GetGame(sessionID)
	if currentGame == nil {
		currentGame = g.NewGame(sessionID)

		g.Store.SaveGame(sessionID, currentGame)
	}

	// bestTime, err := g.GetBestTime(r.Context(), a.DB, sessionID)
	bestTime, err := g.Store.GetBestTime(r.Context(), sessionID)
	if err != nil {
		slog.Error("Ошибка при получении рекорда", "session_id", sessionID, "err", err)
		bestTime = "--:--"
	}

	data := struct {
		*g.Game
		BestRecord string
	}{
		Game:       currentGame,
		BestRecord: bestTime,
	}

	err = tpl.Execute(w, data)
	if err != nil {
		slog.Error("ошибка", "err", err)
	}
}

func (a *App) ClickHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	game, sessionID, err := getGameFromRequest(w, r)
	if err != nil {
		return
	}

	valStr := r.URL.Query().Get("val")
	val, err := strconv.Atoi(valStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "неверный формат числа")
		return
	}

	slog.Info("клик получен", "val", val)

	// if err == nil {
	cr := ClickResponse{
		IsCorrect:  false,
		NextNumber: game.NextNumber,
		Status:     game.Status,
	}

	game.RWmu.Lock()
	if game.Status != "Playing" {
		cr.Status = game.Status
		game.RWmu.Unlock()
		json.NewEncoder(w).Encode(cr)
		return
	}

	if val == game.NextNumber {
		game.NextNumber++

		cr.IsCorrect = true
		cr.NextNumber = game.NextNumber

		if game.NextNumber == 26 {
			game.Status = "Won!"
			cr.Status = "Won!"

			cr.TimeTaken = g.FormatDuration(time.Since(game.StartTime))
			duration := time.Since(game.StartTime)

			log := logger.WithSession(sessionID)
			log.Info("Пользователь нашел все числа")

			a.WG.Add(1)
			go func() {
				defer a.WG.Done()

				ctx := context.Background()
				timeout := 3 * time.Second
				ctx, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				err := g.Store.SaveRecord(ctx, sessionID, duration)
				if err != nil {
					slog.Error("Ошибка при сохранении рекорда", "session_id", sessionID, "error", err)
				}
			}()
		}
	}
	cr.NextNumber = game.NextNumber
	game.RWmu.Unlock()

	if err := json.NewEncoder(w).Encode(cr); err != nil {
		slog.Error("не удалось закодировать", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}
	// }
}

func (a *App) TimerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	game, _, err := getGameFromRequest(w, r)
	if err != nil {
		return
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	limitTimer := time.NewTimer(3 * time.Minute)
	defer limitTimer.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-ticker.C:
			game.RWmu.RLock()
			status := game.Status
			startTime := game.StartTime
			game.RWmu.RUnlock()

			if status != "Playing" {
				return
			}

			timeStr := g.FormatDuration(time.Since(startTime))
			fmt.Fprintf(w, "data: %s\n\n", timeStr)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

		case <-limitTimer.C:
			game.RWmu.Lock()
			game.Status = "Timeout"
			game.RWmu.Unlock()

			fmt.Fprintf(w, "data: Timeout\n\n")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			return
		}
	}
}

func (a *App) RestartHandler(w http.ResponseWriter, r *http.Request) {
	foundGame, sessionID, err := getGameFromRequest(w, r)
	if err != nil {
		return
	}

	foundGame.Reset()

	log := logger.WithSession(sessionID)
	log.Info("Игрок сбросил игру")

	http.Redirect(w, r, "/", http.StatusFound)
}
