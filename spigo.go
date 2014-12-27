// simulate protocol interactions in go - spigo
// terminology is a mix of promise theory and flying spaghetti monster lore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/adrianco/spigo/fsm"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/graphml"
	"github.com/adrianco/spigo/pirate"
	"log"
	"os"
	"time"
)

var Population, duration int
var reload bool

func main() {
	flag.IntVar(&Population, "p", 100, "  Pirate population")
	flag.IntVar(&duration, "d", 10, "   Simulation duration in seconds")
	flag.BoolVar(&graphml.Enabled, "g", false, "Enable GraphML logging")
	flag.BoolVar(&graphjson.Enabled, "j", false, "Enable GraphJSON logging")
	flag.BoolVar(&reload, "r", false, "Reload spigo.json first")
	flag.Parse()
	if graphml.Enabled && graphjson.Enabled {
		fmt.Println("Pick either GraphML or JSON output, not both\n")
		return
	}
	fmt.Println("Spigo population", Population, "pirates")
	noodles := make(map[string]chan gotocol.Message, Population)
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
							fmt.Print(" ",l,":", w)
						}
						fmt.Println()
					default: fmt.Println(i, u)
					}
				}
			default:
				fmt.Println(k, "is of a type I don't know how to handle")
			}
		}
		return
	} else {
		for i := 1; i <= Population; i++ {
			name := fmt.Sprintf("Pirate%d", i)
			noodles[name] = make(chan gotocol.Message)
			go pirate.Listen(noodles[name])
		}
	}
	fsm.ChatSleep = time.Duration(duration) * time.Second
	fsm.Touch(noodles)
}
