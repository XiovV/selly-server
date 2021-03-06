package server

import (
	"encoding/json"
	"errors"
	"github.com/XiovV/selly-server/hub"
	"github.com/XiovV/selly-server/models"
	"github.com/XiovV/selly-server/rabbitmq"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	version = "0.1.1"
)

type Server struct {
	upgrader websocket.Upgrader
	hub      *hub.Hub
	mq       *rabbitmq.RabbitMQ
	log      *zap.SugaredLogger
}

func New(hub *hub.Hub, mq *rabbitmq.RabbitMQ, logger *zap.SugaredLogger) *Server {
	s := &Server{upgrader: websocket.Upgrader{}, hub: hub, mq: mq, log: logger}

	go s.consumeQueue()

	return s
}

func (s *Server) Serve() {
	s.log.Infow("running", "port", os.Getenv("PORT"), "environment", os.Getenv("ENV"), "version", version)

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
				s.log.Error(err)
			case strings.Contains(err.Error(), "websocket: close 1000 (normal)"), strings.Contains(err.Error(), "websocket: close 1006 (abnormal closure)"):
				s.log.Debugw("disconnected", "user", sender, "reason", err)
				s.hub.Pop(sender)
			default:
				s.log.Error("unknown error:", err)
				s.log.Debugw("disconnected", "user", sender, "reason", err)
				s.hub.Pop(sender)
			}
			break
		}

		receiver, exists := s.hub.Get(message.Receiver)

		if !exists {
			s.log.Debugw("receiver does not exist in the local hub, pushing message to RabbitMQ", "sender", message.Sender, "receiver", message.Receiver)
			err = s.mq.Publish(message)
			if err != nil {
				s.log.Error("failed to publish a message:", err)
			}
		} else {
			s.log.Debugw("receiver exists in the local hub, sending message directly to the user", "sender", message.Sender, "receiver", message.Receiver)

			go s.sendMessage(message, receiver)
		}

		s.log.Debugw("sending", "message", message)
	}
}

func (s *Server) sendMessage(msg models.Message, receiver *websocket.Conn) {
	err := receiver.WriteJSON(msg)
	if err != nil {
		s.log.Error("failed to send a message:", err)
	}
}

func (s *Server) consumeQueue() {
	messages, err := s.mq.Consume()
	if err != nil {
		s.log.Error("failed to consume:", err)
	}

	var message models.Message
	for d := range messages {
		d.Ack(false)
		err = json.Unmarshal(d.Body, &message)
		if err != nil {
			s.log.Error("failed to unmarshal message:", err)
		}

		s.log.Debugw("consumed:", "message", message)
		receiver, exists := s.hub.Get(message.Receiver)
		if exists {
			s.log.Debugw("consumed receiver exists in the local hub", "receiver", message.Receiver)
			go s.sendMessage(message, receiver)
		}
	}
}
