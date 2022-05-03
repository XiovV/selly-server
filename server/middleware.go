package server

import (
	"fmt"
	"github.com/XiovV/gob_server/hub"
	"log"
	"net/http"
)

func (s *Server) OnConnect(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		username := params.Get("username")

		r = s.contextSetSender(r, username)

		c, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}

		r = s.contextSetConnection(r, c)

		fmt.Println(username, "connected")

		s.hub.Push(hub.Client{Username: username, Conn: c})

		next(w, r)
	}
}
