// Package main for spigo - simulate protocol interactions in go.
// Terminology is a mix of promise theory and flying spaghetti monster lore
package main

import (
	"flag"
	"fmt"
	"github.com/adrianco/spigo/fsm"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/logger"
	"github.com/adrianco/spigo/pirate"
	"log"
	"time"
)

var arch string
var population, duration int
var reload, msglog bool

// main handles command line flags and starts up an architecture
func main() {
	flag.StringVar(&arch, "a", "fsm", "Architecture to create or read")
	flag.IntVar(&population, "p", 100, "  Pirate population")
	flag.IntVar(&duration, "d", 10, "   Simulation duration in seconds")
	flag.BoolVar(&logger.GraphmlEnabled, "g", false, "Enable GraphML logging of nodes and edges")
	flag.BoolVar(&logger.GraphjsonEnabled, "j", false, "Enable GraphJSON logging of nodes and edges")
	flag.BoolVar(&msglog, "m", false, "Enable console logging of every message")
	flag.BoolVar(&reload, "r", false, "Reload <arch>.json to setup architecture")
	flag.Parse()
	if msglog { // pass on the verbose logging option to all message listeners
		logger.Msglog = true
		fsm.Msglog = true
		pirate.Msglog = true
	}
	noodles := make(map[string]chan gotocol.Message, population)
	if reload {
		log.Println("Spigo reloading from " + arch + ".json")
		g := graphjson.ReadArch(arch)
		for _, element := range g.Graph {
			if element.Node != "" {
				fmt.Println("Create " + element.Service + " " + element.Node)
			}
			if element.Edge != "" {
				fmt.Println("Link " + element.Source + " > " + element.Target)
			}
		}
		return
	} else {
		log.Println("Spigo: population", population, "pirates")
		for i := 1; i <= population; i++ {
			name := fmt.Sprintf("Pirate%d", i)
			noodles[name] = make(chan gotocol.Message)
			go pirate.Start(noodles[name])
		}
	}
	// start up the selected architecture
	switch arch {
	case "fsm":
		if logger.GraphjsonEnabled || logger.GraphmlEnabled {
			go logger.Start("fsm") // start logger first
		}
		fsm.ChatSleep = time.Duration(duration) * time.Second
		fsm.Touch(noodles)
		log.Println("spigo: fsm complete")
		if logger.Logchan != nil {
			for {
				log.Printf("Logger has %v messages left to flush\n", len(logger.Logchan))
				time.Sleep(time.Second)
				if len(logger.Logchan) == 0 {
					break
				}
			}
		}
	default:
		log.Fatal("Architecture " + arch + " isn't recognized")
	}
}
