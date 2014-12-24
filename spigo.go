// simulate protocol interactions in go - spigo
// terminology is a mix of promise theory and flying spaghetti monster lore

package main

import (
	"flag"
	"fmt"
	"github.com/adrianco/spigo/fsm"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/graphml"
	"github.com/adrianco/spigo/pirate"
	"time"
)

var Population, duration int

func main() {
	flag.IntVar(&Population, "p", 100, "Pirate population")
	flag.IntVar(&duration, "d", 10, "Simulation duration in seconds")
	flag.BoolVar(&graphml.Enabled, "g", false, "Enable GraphML logging")
	flag.BoolVar(&graphjson.Enabled, "j", false, "Enable GraphJSON logging")
	flag.Parse()
	if graphml.Enabled && graphjson.Enabled {
		fmt.Println("Pick either GraphML or JSON output, not both\n")
		return
	}
	fmt.Println("Spigo population", Population, "pirates")
	noodles := make(map[string]chan gotocol.Message, Population)
	for i := 1; i <= Population; i++ {
		name := fmt.Sprintf("Pirate%d", i)
		noodles[name] = make(chan gotocol.Message)
		go pirate.Listen(noodles[name])
	}
	fsm.ChatSleep = time.Duration(duration) * time.Second
	fsm.Touch(noodles)
}
