package main

import (
	"flag"
	"strings"
	"fmt"
	"github.com/HearthSim/stove/bnet"
	"github.com/HearthSim/stove/pegasus"
)

const (
	CONN_DEFAULT_HOST = "localhost"
	CONN_DEFAULT_PORT = 1119
)

func main() {
	addr := fmt.Sprintf("%s:%d", CONN_DEFAULT_HOST, CONN_DEFAULT_PORT)
	flag.StringVar(&addr, "bind", addr, "The address to run on")
	flag.Parse()

	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf("%s:%d", addr, CONN_DEFAULT_PORT)
	}

	serv := bnet.NewServer()
	serv.RegisterGameServer("WTCG", pegasus.NewServer(serv))

	fmt.Printf("Listening on %s ...\n", addr)
	err := serv.ListenAndServe(addr)
	if err != nil {
		fmt.Println(err)
	}
}
