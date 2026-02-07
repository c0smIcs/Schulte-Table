package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/c0smIcs/SchulteTable/internal/game"
	"github.com/c0smIcs/SchulteTable/internal/handler"
)

func main() {
    dsn := "host=localhost user=postgres password=1212 dbname=schulte port=5432 sslmode=disable"
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

	fmt.Println("Сервер запущен по адресу: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
