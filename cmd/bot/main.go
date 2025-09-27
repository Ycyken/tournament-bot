package main

import (
	"log"
	"os"

	"github.com/Ycyken/tournament-bot/internal/bot"
	"github.com/Ycyken/tournament-bot/internal/service"
	"github.com/Ycyken/tournament-bot/internal/store/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	store, err := postgres.NewPostgresStore(dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	svc := service.New(store)

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	bt, _ := bot.NewBot(svc, token)
	bt.Run()

}
