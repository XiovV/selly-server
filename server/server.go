package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/XiovV/gob_server/hub"
	"github.com/XiovV/gob_server/models"
	"github.com/XiovV/gob_server/rabbitmq"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Server struct {
	upgrader websocket.Upgrader
	hub      *hub.Hub
	mq       *rabbitmq.RabbitMQ
}

func New(hub *hub.Hub, mq *rabbitmq.RabbitMQ) *Server {
	s := &Server{upgrader: websocket.Upgrader{}, hub: hub, mq: mq}

	go s.consumeQueue()

	return s
}

func (s *Server) Serve() {
	fmt.Println("running on", os.Getenv("PORT"))
	http.HandleFunc("/chat", s.OnConnect(s.Chat))

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

func (s *Server) Chat(w http.ResponseWriter, r *http.Request) {
	c := s.contextGetConnection(r)
	defer c.Close()

	sender := s.contextGetSender(r)

	var message models.Message
	for {
		err := c.ReadJSON(&message)
		if err != nil {
			switch {
			case errors.Is(err, io.ErrUnexpectedEOF):
				fmt.Println("unexpected eof")
			case strings.Contains(err.Error(), "websocket: close 1000 (normal)"), strings.Contains(err.Error(), "websocket: close 1006 (abnormal closure)"):
				fmt.Printf("%s disconnected: %s\n", sender, err)
				s.hub.Pop(sender)
			default:
				fmt.Println("unknown error:", err)
				fmt.Println(sender, "disconnected")
				s.hub.Pop(sender)
			}
			break
		}

		receiver, exists := s.hub.Get(message.Receiver)

		if !exists {
			err = s.mq.Publish(message)
			if err != nil {
				fmt.Println("failed to publish a message")
			}
		} else {
			receiver.WriteJSON(message)
		}

		fmt.Printf("SENDING: %+v\n", message)
	}
}

func (s *Server) consumeQueue() {
	msgs, err := s.mq.Consume()
	if err != nil {
		fmt.Println("failed to consume:", err)
	}

	var message models.Message
	for d := range msgs {
		d.Ack(false)
		err = json.Unmarshal(d.Body, &message)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("RECEIVED: %+v\n", message)
		receiver, exists := s.hub.Get(message.Receiver)
		if exists {
			receiver.WriteJSON(message)
		}
	}
}
