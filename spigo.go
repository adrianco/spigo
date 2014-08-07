// simulate protocol interactions in go - spigo
// terminology is a mix of promise theory and flying spaghetti monster lore

package main

import (
	"fmt"
	"flag"
	"time"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/fsm"
	"github.com/adrianco/spigo/pirate"
	"github.com/adrianco/spigo/graphml"
)

var Population, duration int

func main() {
	flag.IntVar(&Population,"p", 100, "Pirate population")
	flag.IntVar(&duration,"d", 10, "Simulation duration in seconds")
	flag.BoolVar(&graphml.Enabled, "g", false, "Enable GraphML logging")
	flag.Parse()
	fmt.Println("Spigo population", Population, "pirates")
	graphml.Setup()
	noodles := make(map[string]chan gotocol.Message, Population)
	for i := 1; i <= Population; i++ {
		name := fmt.Sprintf("Pirate%d", i)
		graphml.Node(name)
		noodles[name] = make(chan gotocol.Message)
		go pirate.Listen(noodles[name])
	}
	fsm.ChatSleep = time.Duration(duration) * time.Second
	fsm.Touch(noodles)
	graphml.Close()
}
