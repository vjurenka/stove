package main

import (
	"github.com/HearthSim/stove/bnet"
)

func main() {
	bnet.NewServer().ListenAndServe("localhost:1119")
}
