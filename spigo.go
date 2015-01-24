// Package main for spigo - simulate protocol interactions in go.
// Terminology is a mix of promise theory and flying spaghetti monster lore
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/adrianco/spigo/fsm"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/graphml"
	"github.com/adrianco/spigo/logger"
	"github.com/adrianco/spigo/pirate"
	"log"
	"os"
	"time"
)

var arch string
var population, duration int
var reload, msglog bool

// main handles command line flags and starts up an architecture
func main() {
	flag.StringVar(&arch, "a", "fsm", "Architecture to create")
	flag.IntVar(&population, "p", 100, "  Pirate population")
	flag.IntVar(&duration, "d", 10, "   Simulation duration in seconds")
	flag.BoolVar(&graphml.Enabled, "g", false, "Enable GraphML logging of nodes and edges")
	flag.BoolVar(&graphjson.Enabled, "j", false, "Enable GraphJSON logging of nodes and edges")
	flag.BoolVar(&msglog, "m", false, "Enable console logging of every message")
	flag.BoolVar(&reload, "r", false, "Reload spigo.json to setup architecture")
	flag.Parse()
	if graphml.Enabled && graphjson.Enabled {
		log.Fatal("Pick either GraphML or JSON output, not both")
	}
	if msglog { // pass on the verbose logging option to all message listeners
		logger.Msglog = true
		fsm.Msglog = true
		pirate.Msglog = true
	}
	log.Println("Spigo: population", population, "pirates")
	noodles := make(map[string]chan gotocol.Message, population)
	if reload {
		file, err := os.Open("spigo.json")
		if err != nil {
			log.Fatal(err)
		}
		dec := json.NewDecoder(file)
		var f interface{}
		if err := dec.Decode(&f); err != nil {
			log.Fatal(err)
		}
		m := f.(map[string]interface{})
		for k, v := range m {
			switch vv := v.(type) {
			case string:
				fmt.Println(k, " ", vv)
			case int:
				fmt.Println(k, "=", vv)
			case []interface{}:
				fmt.Println(k, "is an array:")
				for i, u := range vv {
					switch uu := u.(type) {
					case map[string]interface{}:
						for l, w := range uu {
							fmt.Print(" ", l, ":", w)
						}
						fmt.Println()
					default:
						fmt.Println(i, u)
					}
				}
			default:
				log.Println(k, "is of a type I don't know how to handle")
			}
		}
		return
	} else {
		for i := 1; i <= population; i++ {
			name := fmt.Sprintf("Pirate%d", i)
			noodles[name] = make(chan gotocol.Message)
			go pirate.Start(noodles[name])
		}
	}
	// start up the selected architecture
	switch arch {
	case "fsm":
		if graphjson.Enabled || graphml.Enabled {
			go logger.Start("fsm") // start logger first
		}
		fsm.ChatSleep = time.Duration(duration) * time.Second
		fsm.Touch(noodles)
		log.Println("spigo: fsm complete")
		if graphjson.Enabled || graphml.Enabled {
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
