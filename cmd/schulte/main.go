package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/c0smIcs/SchulteTable/internal/game"
	"github.com/c0smIcs/SchulteTable/internal/handler"
	"github.com/c0smIcs/SchulteTable/internal/logger"
	"github.com/spf13/viper"
)

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

func main() {
	if err := initConfig(); err != nil {
		log.Fatalf("error initizlizing configs: %s", err.Error())
	}

	logger.InitLogger()

	// БД:
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		viper.GetString("db.user"),
		viper.GetString("db.password"),
		viper.GetString("db.host"),
		viper.GetString("db.port"),
		viper.GetString("db.dbname"),
		viper.GetString("db.sslmode"),
	)

	pool, err := game.InitDB(dsn)
	if err != nil {
		os.Exit(1)
	}
	defer pool.Close()

	wg := &sync.WaitGroup{}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app := &handler.App{
		DB: pool,
		WG: wg,
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	fsJS := http.FileServer(http.Dir("./ui/js"))
	mux.Handle("/js/", http.StripPrefix("/js/", fsJS))

	mux.HandleFunc("/", app.IndexHandler)
	mux.HandleFunc("/click", app.ClickHandler)
	mux.HandleFunc("/timer", app.TimerHandler)
	mux.HandleFunc("/restart", app.RestartHandler)

	// port := os.Getenv("SERVER_PORT")
	port := viper.GetString("serverport")
	if port == "" {
		slog.Error("Ошибка с портом", "порт пустой", port)
		return
	}
	addr := ":" + port

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	slog.Info("Сервер запущен",
		slog.String("port", port),
		slog.String("url", "http://localhost:8080"),
	)

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("Сервер аварийно остановился", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()
	slog.Info("Получен сигнал завершения, начинаем остановку...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Warn("Ошибка при остановке HTTP сервера", "err", err)
	}
	app.WG.Wait()
}
