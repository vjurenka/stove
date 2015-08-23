package main

import (
	"flag"
	"fmt"
	"github.com/HearthSim/stove/bnet"
	"github.com/HearthSim/stove/pegasus"
	_ "github.com/rakyll/gom/http"
	"net/http"
	"strings"
)

const (
	CONN_DEFAULT_HOST = "localhost"
	CONN_DEFAULT_PORT = 1119
)

func main() {
	addr := fmt.Sprintf("%s:%d", CONN_DEFAULT_HOST, CONN_DEFAULT_PORT)
	flag.StringVar(&addr, "bind", addr, "The address to run on")
	runMigrate := flag.Bool("migrate", false, "Perform a database migration and exit")
	flag.Parse()

	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf("%s:%d", addr, CONN_DEFAULT_PORT)
	}

	if *runMigrate {
		fmt.Printf("Performing database migration\n")
		pegasus.Migrate()
		return
	}

	go func() {
		httpAddr := "localhost:6060"
		fmt.Printf("Debug http server listening on %s ...\n", httpAddr)
		fmt.Println(http.ListenAndServe(httpAddr, nil))
	}()

	serv := bnet.NewServer()
	serv.RegisterGameServer("WTCG", pegasus.NewServer(serv))

	fmt.Printf("Listening on %s ...\n", addr)
	err := serv.ListenAndServe(addr)
	if err != nil {
		fmt.Println(err)
	}
}
