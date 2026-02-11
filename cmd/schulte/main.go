package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/c0smIcs/SchulteTable/internal/game"
	"github.com/c0smIcs/SchulteTable/internal/handler"
	"github.com/c0smIcs/SchulteTable/internal/logger"
	"github.com/joho/godotenv"
)

func main() {
	logger.InitLogger()

	err := godotenv.Load()
	if err != nil {
		slog.Error("Ошибка при загрузки .env файла", slog.Any("err", err))
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("HOST"),
		os.Getenv("USER"),
		os.Getenv("PASSWORD"),
		os.Getenv("DBNAME"),
		os.Getenv("PORT"),
		os.Getenv("SSLMODE"),
	)

	dsnM := fmt.Sprintf(
		"host=%s user=%s dbname=%s port=%s",
		os.Getenv("HOST"),
		os.Getenv("USER"),
		os.Getenv("DBNAME"),
		os.Getenv("PORT"),
	)

	db, err := game.InitDB(dsn)
	if err != nil {
		logger.LoggerDBError(err, dsnM)
		os.Exit(1)
	}
	logger.LoggerDBConnect(dsnM)

	app := &handler.App{
		DB: db,
	}

	fs := http.FileServer(http.Dir("./ui/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	fsJS := http.FileServer(http.Dir("./ui/js"))
	http.Handle("/js/", http.StripPrefix("/js/", fsJS))

	http.HandleFunc("/", app.IndexHandler)
	http.HandleFunc("/click", app.ClickHandler)
	http.HandleFunc("/timer", app.TimerHandler)
	http.HandleFunc("/restart", app.RestartHandler)

	port := os.Getenv("SERVER_PORT")
	addr := ":" + port

	slog.Info("Сервер запущен", slog.String("port", port),
		slog.String("", "http://localhost:8080"))
	
	err = http.ListenAndServe(addr, nil); if err != nil {
		slog.Error("Сервер аварийно остановился", slog.String("error", err.Error()))
	}
}
