package main

import (
	"fmt"
	"net/http"

	"github.com/c0smIcs/SchulteTable/internal/handler"
)

func main() {
	fs := http.FileServer(http.Dir("./ui/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fsJS := http.FileServer(http.Dir("./ui/js"))
	http.Handle("/js/", http.StripPrefix("/js/", fsJS))

	http.HandleFunc("/", handler.IndexHandler)
	http.HandleFunc("/click", handler.ClickHandler)
	http.HandleFunc("/timer", handler.TimerHandler)
	http.HandleFunc("/restart", handler.RestartHandler)

	fmt.Println("Сервер запущен по адресу: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
