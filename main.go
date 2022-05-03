package main

import (
	"github.com/XiovV/selly-server/hub"
	"github.com/XiovV/selly-server/rabbitmq"
	"github.com/XiovV/selly-server/server"
	"go.uber.org/zap"
	"log"
	"os"
)

func main() {
	checkEnvVars()

	var logger *zap.Logger
	if os.Getenv("ENV") == "DEV" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	hub := hub.New()
	rabbitmq, err := rabbitmq.New()
	if err != nil {
		log.Fatal(err)
	}

	s := server.New(hub, rabbitmq, sugar)
	s.Serve()
}
