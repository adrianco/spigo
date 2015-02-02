// Package main for spigo - simulate protocol interactions in go.
// Terminology is a mix of promise theory and flying spaghetti monster lore
package main

import (
	"flag"
	"github.com/adrianco/spigo/edda"
	"github.com/adrianco/spigo/fsm"
	"github.com/adrianco/spigo/netflixoss"
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
	flag.BoolVar(&edda.GraphmlEnabled, "g", false, "Enable GraphML logging of nodes and edges")
	flag.BoolVar(&edda.GraphjsonEnabled, "j", false, "Enable GraphJSON logging of nodes and edges")
	flag.BoolVar(&msglog, "m", false, "Enable console logging of every message")
	flag.BoolVar(&reload, "r", false, "Reload <arch>.json to setup architecture")
	flag.Parse()
	if msglog { // pass on the verbose logging option to all message listeners
		edda.Msglog = true
		fsm.Msglog = true
		netflixoss.Msglog = true
	}
	// start up the selected architecture
	switch arch {
	case "fsm":
		if edda.GraphjsonEnabled || edda.GraphmlEnabled {
			go edda.Start(arch) // start edda first
		}
		fsm.ChatSleep = time.Duration(duration) * time.Second
		if reload {
			fsm.Reload(arch)
		} else {
			fsm.Start()
		}
		log.Println("spigo: fsm complete")
		if edda.Logchan != nil {
			for {
				log.Printf("Logger has %v messages left to flush\n", len(edda.Logchan))
				time.Sleep(time.Second)
				if len(edda.Logchan) == 0 {
					break
				}
			}
		}
	case "netflixoss":
		if edda.GraphjsonEnabled || edda.GraphmlEnabled {
			go edda.Start(arch) // start edda first
		}
		netflixoss.RunSleep = time.Duration(duration) * time.Second
		netflixoss.Population = fsm.Population
		if reload {
			netflixoss.Reload(arch)
		} else {
			netflixoss.Start()
		}
		log.Println("spigo: netflixoss complete")
		if edda.Logchan != nil {
			for {
				log.Printf("Logger has %v messages left to flush\n", len(edda.Logchan))
				time.Sleep(time.Second)
				if len(edda.Logchan) == 0 {
					break
				}
			}
		}
	default:
		log.Fatal("Architecture " + arch + " isn't recognized")
	}
}
