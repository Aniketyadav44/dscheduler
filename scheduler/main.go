package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/Aniketyadav44/dscheduler/scheduler/internal/config"
	"github.com/Aniketyadav44/dscheduler/scheduler/internal/services"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	log.Println("Starting scheduler service...")

	cfg, err := config.LoadConfig()
	if err != nil {
		stop()
		log.Fatal("error in loading config: ", err.Error())
	}

	// batch processor for fetching jobs from db every 10 mins
	// and adding them to ZSET
	batchProcessor := services.NewBatchProcessor(cfg.DB, cfg.Redis)
	batchProcessor.Start()
	defer batchProcessor.Stop()

	// starting poller ticking, which gets job of every min from ZSET
	// and publishes them to redis stream
	publisher := services.NewPublisher(cfg.Redis)
	publisher.Start()
	defer publisher.Stop()

	<-ctx.Done()
	stop()
	log.Println("Stopping scheduler service...")
}
