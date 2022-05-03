package server

import (
	"context"
	"github.com/gorilla/websocket"
	"net/http"
)

const (
	connectionContextKey = "wsConnection"
	senderContextKey     = "sender"
)

func (s *Server) contextSetConnection(r *http.Request, conn *websocket.Conn) *http.Request {
	ctx := context.WithValue(r.Context(), connectionContextKey, conn)
	return r.WithContext(ctx)
}

func (s *Server) contextGetConnection(r *http.Request) *websocket.Conn {
	conn, ok := r.Context().Value(connectionContextKey).(*websocket.Conn)
	if !ok {
		panic("missing connection in request context")
	}

	return conn
}

func (s *Server) contextSetSender(r *http.Request, sender string) *http.Request {
	ctx := context.WithValue(r.Context(), senderContextKey, sender)
	return r.WithContext(ctx)
}

func (s *Server) contextGetSender(r *http.Request) string {
	conn, ok := r.Context().Value(senderContextKey).(string)
	if !ok {
		panic("missing sender in request context")
	}

	return conn
}
