package main

import (
	"fmt"
	"github.com/HearthSim/stove/bnet"
	"github.com/HearthSim/stove/config"
	"github.com/HearthSim/stove/pegasus"
	_ "github.com/rakyll/gom/http"
	"log"
	"net/http"
	"strings"
)

func main() {
	addr := config.Config.ListenAddress
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf("%s:%d", addr, 1119)
	}

	if config.Config.Migrate {
		fmt.Printf("Performing database migration\n")
		pegasus.Migrate()
		return
	}

	debugListen := config.Config.DebugListenAddress
	if len(debugListen) != 0 {
		go func() {
			log.Printf("Debug http server listening on %s ...\n", debugListen)
			log.Println(http.ListenAndServe(debugListen, nil))
		}()
	}

	serv := bnet.NewServer()
	serv.RegisterGameServer("WTCG", pegasus.NewServer(serv))

	log.Printf("Listening on %s ...\n", addr)
	err := serv.ListenAndServe(addr)
	if err != nil {
		log.Println(err)
	}
}
