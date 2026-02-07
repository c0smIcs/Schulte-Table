package main

import (
	"fmt"
	"log"
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
		log.Fatalf("Ошибка при загрузки .env файла")
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

	db, err := game.InitDB(dsn)
	if err != nil {
		log.Fatal("Не удалось подключиться к БД: ", err)
	}

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

	logger.StartServer()
	http.ListenAndServe(":8080", nil)
}
