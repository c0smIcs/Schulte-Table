package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	g "github.com/c0smIcs/SchulteTable/internal/game"
	"github.com/c0smIcs/SchulteTable/internal/logger"
	"gorm.io/gorm"
)

type ClickResponse struct {
	IsCorrect  bool   `json:"is_correct"`
	NextNumber int    `json:"next_number"`
	Status     string `json:"status"`
	TimeTaken  string `json:"time_taken"`
	BestRecord string `json:"best_record"`
}

type App struct {
	DB *gorm.DB
}

var tpl = template.Must(template.ParseFiles("ui/html/index.html"))

func getGameFromRequest(w http.ResponseWriter, r *http.Request) (*g.Game, string, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "сессия не найдена", http.StatusForbidden)
		return nil, "", err
	}

	return g.Store.GetGame(cookie.Value), cookie.Value, nil
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

	bestTime, err := g.GetBestTime(r.Context(), a.DB, sessionID)
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

	tpl.Execute(w, data)
}

func (a *App) ClickHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	game, sessionID, err := getGameFromRequest(w, r)
	if err != nil {
		return
	}

	valStr := r.URL.Query().Get("val")

	val, err := strconv.Atoi(valStr)
	if err == nil {
		cr := ClickResponse{
			IsCorrect:  false,
			NextNumber: game.NextNumber,
			Status:     game.Status,
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

				go func() {
					err := g.SaveRecord(context.Background(), a.DB, sessionID, duration)
					if err != nil {
						slog.Error("Ошибка при сохранении рекорда", "session_id", sessionID, "error", err)
					}
				}()
			}
		}

		json.NewEncoder(w).Encode(cr)
	}
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

	for {
		select {
		case <-r.Context().Done():
			return

		case <-ticker.C:
			if game.Status == "Won!" {
				return
			}

			timeStr := g.FormatDuration(time.Since(game.StartTime))

			fmt.Fprintf(w, "data: %s\n\n", timeStr)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
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
