package game

import (
	"log"
	"net"
)

// Handles client connections and hands them off to the appropriate game
// instance
type server struct {
	games       []*Game
	sock        net.Listener
	gameHandles map[int32]*Game
}

func NewServer(listenAddr string) *server {
	log.Printf("game server listening on %s", listenAddr)
	res := &server{}
	res.games = []*Game{}
	res.gameHandles = map[int32]*Game{}
	var err error
	res.sock, err = net.Listen("tcp", listenAddr)
	if err != nil {
		panic(err)
	}
	go res.serve()
	return res
}

func (s *server) serve() {
	for {
		c, err := s.sock.Accept()
		if err != nil {
			log.Printf("game server: error in accept: %v", err)
			continue
		}
		sess := NewSession(s, c)
		go sess.handle()
	}
}

func (s *server) GameFromHandle(h int32) *Game {
	if game, ok := s.gameHandles[h]; ok {
		return game
	} else {
		log.Panicf("handshake game handle %d not found", h)
		return nil
	}
}

var gameServer = NewServer(":1120")
