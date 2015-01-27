// Package main for spigo - simulate protocol interactions in go.
// Terminology is a mix of promise theory and flying spaghetti monster lore
package main

import (
	"flag"
	"github.com/adrianco/spigo/fsm"
	"github.com/adrianco/spigo/logger"
	"log"
	"time"
)

var arch string
var reload, msglog bool
var duration int

// main handles command line flags and starts up an architecture
func main() {
	flag.StringVar(&arch, "a", "fsm", "Architecture to create or read")
	flag.IntVar(&fsm.Population, "p", 100, "  Pirate population")
	flag.IntVar(&duration, "d", 10, "   Simulation duration in seconds")
	flag.BoolVar(&logger.GraphmlEnabled, "g", false, "Enable GraphML logging of nodes and edges")
	flag.BoolVar(&logger.GraphjsonEnabled, "j", false, "Enable GraphJSON logging of nodes and edges")
	flag.BoolVar(&msglog, "m", false, "Enable console logging of every message")
	flag.BoolVar(&reload, "r", false, "Reload <arch>.json to setup architecture")
	flag.Parse()
	if msglog { // pass on the verbose logging option to all message listeners
		logger.Msglog = true
		fsm.Msglog = true
	}
	// start up the selected architecture
	switch arch {
	case "fsm":
		if logger.GraphjsonEnabled || logger.GraphmlEnabled {
			go logger.Start("fsm") // start logger first
		}
		fsm.ChatSleep = time.Duration(duration) * time.Second
		fsm.Start(reload) // tell fsm to reload or create new pirates
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
