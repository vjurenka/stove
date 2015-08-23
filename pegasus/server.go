package pegasus

import (
	"github.com/HearthSim/stove/bnet"
)

type Server struct {
	host *bnet.Server
}

func NewServer(serv *bnet.Server) *Server {
	res := &Server{}
	return res
}

func (s *Server) Connect(sess *bnet.Session) {
	BindSession(s, sess)
}
