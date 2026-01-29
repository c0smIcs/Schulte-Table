package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/c0smIcs/SchulteTable/internal/game"
)

type ClickResponse struct {
	IsCorrect  bool   `json:"is_correct"`
	NextNumber int    `json:"next_number"`
	Status     string `json:"status"`
	TimeTaken  string `json:"time_taken"`
}

var tpl = template.Must(template.ParseFiles("ui/html/index.html"))

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	var sessionID string

	if err != nil {
		sessionID = game.GenerateSessionID()

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
		})
	} else {
		sessionID = cookie.Value
	}

	game := game.Store.GetGame(sessionID)

	tpl.Execute(w, game)
}

func ClickHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "сессия не найдена", http.StatusForbidden)
		return
	}

	sessionID := cookie.Value

	game := game.Store.GetGame(sessionID)

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

				duration := time.Since(game.StartTime)
				minutes := int(duration.Minutes())
				seconds := int(duration.Seconds())

				cr.TimeTaken = fmt.Sprintf("%02d:%02d", minutes, seconds)
			}
		}

		json.NewEncoder(w).Encode(cr)
	}
}

func TimerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "сессия не найдена", http.StatusForbidden)
		return
	}
	
	sessionID := cookie.Value

	game := game.Store.GetGame(sessionID)

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

			duration := time.Since(game.StartTime)

			minutes := int(duration.Minutes()) % 60
			seconds := int(duration.Seconds()) % 60
			milliseconds := duration.Milliseconds() / 100 % 10

			timeStr := fmt.Sprintf("%02d:%02d:%02d", minutes, seconds, milliseconds)

			fmt.Fprintf(w, "data: %s\n\n", timeStr)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

func RestartHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "сессия не найден", http.StatusForbidden)
		return
	}

	sessionID := cookie.Value

	foundGame := game.Store.GetGame(sessionID)

	matrix := game.GenerateBoard()

	foundGame.NextNumber = 1
	foundGame.Status = "Playing"
	foundGame.StartTime = time.Now()
	foundGame.Board = matrix

	http.Redirect(w, r, "/", http.StatusFound)
}