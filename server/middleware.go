package server

import (
	"github.com/XiovV/selly-server/hub"
	"github.com/XiovV/selly-server/jwt"
	"log"
	"net/http"
)

func (s *Server) OnConnect(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		token := params.Get("jwt")

		validatedToken, err := jwt.Validate(token)
		if err != nil {
			s.log.Error("error validating jwt:", err)
			return
		}

		sellyID := jwt.GetClaimString(validatedToken, "sellyId")

		r = s.contextSetSender(r, sellyID)

		c, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}

		r = s.contextSetConnection(r, c)

		s.log.Debugw("connected", "user", sellyID)

		s.hub.Push(hub.Client{SellyID: sellyID, Conn: c})

		next(w, r)
	}
}
