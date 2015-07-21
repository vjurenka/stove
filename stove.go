package main

import (
	"fmt"
	"github.com/HearthSim/stove/bnet"
	"github.com/HearthSim/stove/pegasus"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = 1119
)

func main() {
	serv := bnet.NewServer()
	serv.RegisterGameServer("WTCG", pegasus.NewServer(serv))

	addr := fmt.Sprintf("%s:%d", CONN_HOST, CONN_PORT)
	fmt.Printf("Listening on %s ...\n", addr)
	serv.ListenAndServe(addr)
}
