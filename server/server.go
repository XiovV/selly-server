package server

import (
	"encoding/json"
	"errors"
	"github.com/XiovV/selly-server/hub"
	"github.com/XiovV/selly-server/models"
	"github.com/XiovV/selly-server/rabbitmq"
	"github.com/XiovV/selly-server/redis"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	version     = "0.1.1"
	messageType = "message"
)

type Server struct {
	upgrader websocket.Upgrader
	hub      *hub.Hub
	mq       *rabbitmq.RabbitMQ
	redis    *redis.Redis
	log      *zap.SugaredLogger
}

func New(hub *hub.Hub, mq *rabbitmq.RabbitMQ, redis *redis.Redis, logger *zap.SugaredLogger) *Server {
	s := &Server{upgrader: websocket.Upgrader{}, hub: hub, mq: mq, redis: redis, log: logger}

	go s.consumeQueue()

	return s
}

type Envelope struct {
	Type string
	Msg  any
}

func (s *Server) Serve() {
	s.log.Infow("running", "port", os.Getenv("PORT"), "environment", os.Getenv("ENV"), "version", version)

	http.HandleFunc("/chat", s.OnConnect(s.Chat))
	http.HandleFunc("/health", s.Health)

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	_, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Error("upgrader error:", err)
		return
	}
}

func (s *Server) Chat(w http.ResponseWriter, r *http.Request) {
	c := s.contextGetConnection(r)
	defer c.Close()

	sender := s.contextGetSender(r)

	var msg json.RawMessage
	payload := Envelope{Msg: &msg}

	for {
		err := c.ReadJSON(&payload)
		if err != nil {
			switch {
			case errors.Is(err, io.ErrUnexpectedEOF):
				s.log.Error(err)
			case strings.Contains(err.Error(), "websocket: close 1000 (normal)"), strings.Contains(err.Error(), "websocket: close 1006 (abnormal closure)"):
				s.log.Debugw("disconnected", "user", sender, "reason", err)
				s.redis.DelUser(sender)
				s.hub.Pop(sender)
			default:
				s.log.Error("unknown error:", err)
				s.log.Debugw("disconnected", "user", sender, "reason", err)
				s.redis.DelUser(sender)
				s.hub.Pop(sender)
			}
			break
		}

		switch payload.Type {
		case messageType:
			s.handleIncomingMessage(msg)
		default:
			log.Fatal("unknown message type")
		}

	}
}

func (s *Server) handleIncomingMessage(msg json.RawMessage) {
	var message models.Message

	if err := json.Unmarshal(msg, &message); err != nil {
		s.log.Error("couldn't unmarshal message:", err)
		return
	}

	message.DateCrated = time.Now().Unix()

	receiver, exists := s.hub.Get(message.Receiver)

	if exists {
		s.log.Debugw("receiver exists in the local hub, sending message directly to the user", "sender", message.Sender, "receiver", message.Receiver)

		go s.sendMessage(message, receiver)
	} else if s.redis.IsUserOnline(message.Receiver) {
		s.log.Debugw("receiver does not exist in the local hub but is online, pushing message to RabbitMQ", "sender", message.Sender, "receiver", message.Receiver)
		err := s.mq.Publish(message)
		if err != nil {
			s.log.Error("failed to publish a message:", err)
		}
	} else {
		s.log.Debugw("user is offline, pushing message to redis", "sender", message.Sender, "receiver", message.Receiver)

		s.redis.PushMessage(message)
	}

	s.log.Debugw("sending", "message", message)
}

func (s *Server) sendMessage(msg models.Message, receiver *websocket.Conn) {
	var m struct {
		Type string
		Msg  models.Message
	}

	m.Type = "message"
	m.Msg = msg

	err := receiver.WriteJSON(m)
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
