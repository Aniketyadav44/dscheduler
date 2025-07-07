package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/Aniketyadav44/dscheduler/worker/internal/config"
	"github.com/Aniketyadav44/dscheduler/worker/internal/services"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("error in loading config: ", err.Error())
	}

	dbService := services.NewDBService(cfg.DB)
	consumer := services.NewConsumer(dbService, cfg.Redis)
	defer consumer.Stop()
	consumer.Start()

	<-ctx.Done()
	stop()
}
