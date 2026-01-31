package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	g "github.com/c0smIcs/SchulteTable/internal/game"
)

type ClickResponse struct {
	IsCorrect  bool   `json:"is_correct"`
	NextNumber int    `json:"next_number"`
	Status     string `json:"status"`
	TimeTaken  string `json:"time_taken"`
}

var tpl = template.Must(template.ParseFiles("ui/html/index.html"))

func getGameFromRequest(w http.ResponseWriter, r *http.Request) (*g.Game, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "сессия не найдена", http.StatusForbidden)
		return nil, err
	}

	return g.Store.GetGame(cookie.Value), nil
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	var sessionID string

	if err != nil {
		sessionID = g.GenerateSessionID()

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
		})
	} else {
		sessionID = cookie.Value
	}

	g := g.Store.GetGame(sessionID)

	tpl.Execute(w, g)
}

func ClickHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	game, err := getGameFromRequest(w, r)
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
			}
		}

		json.NewEncoder(w).Encode(cr)
	}
}

func TimerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	game, err := getGameFromRequest(w, r)
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

func RestartHandler(w http.ResponseWriter, r *http.Request) {
	foundGame, err := getGameFromRequest(w, r)
	if err != nil {
		return
	}

	foundGame.Reset()

	http.Redirect(w, r, "/", http.StatusFound)
}
