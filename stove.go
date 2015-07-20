package main

import (
	"github.com/HearthSim/stove/bnet"
	"github.com/HearthSim/stove/pegasus"
)

func main() {
	serv := bnet.NewServer()
	serv.RegisterGameServer("WTCG", pegasus.NewServer(serv))
	serv.ListenAndServe("localhost:1119")
}
