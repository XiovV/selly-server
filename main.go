package main

import (
	"github.com/XiovV/gob_server/hub"
	"github.com/XiovV/gob_server/rabbitmq"
	"github.com/XiovV/gob_server/server"
	"log"
)

func main() {
	hub := hub.New()
	rabbitmq, err := rabbitmq.New()
	if err != nil {
		log.Fatal(err)
	}

	s := server.New(hub, rabbitmq)
	s.Serve()
}
