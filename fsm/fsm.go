// Package fsm implements a Flexible State Manager (Flying Spaghetti Monster)
// It creates and controls a large social network of pirates via channels (the noodly touch)
// or reads in a network from a json file. I also logs the architecture (nodes and links) as it evolves
package fsm

import (
	"fmt"
	"github.com/adrianco/spigo/edda"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/pirate"
	"log"
	"math/rand"
	"time"
)

// Population count of pirates to create
var Population int

// ChatSleep duration is set via command line flag to tell fsm how long to let pirates chat
var ChatSleep time.Duration

// Msglog toggles whether to log every message received to the console
var Msglog bool

// noodles channels mapped by pirate name connects fsm to everyone
var noodles map[string]chan gotocol.Message
var names []string
var listener chan gotocol.Message

// Reload the network from a file
func Reload(arch string) {
	pirate.Msglog = Msglog                // pass on console message log flag if set
	listener = make(chan gotocol.Message) // listener for fsm
	log.Println("fsm reloading from " + arch + ".json")
	g := graphjson.ReadArch(arch)
	Population = 0 // just to make sure
	// count how many nodes there are
	for _, element := range g.Graph {
		if element.Node != "" {
			Population++
		}
	}
	// create the map of channels
	noodles = make(map[string]chan gotocol.Message, Population)
	// Start all the services
	for _, element := range g.Graph {
		if element.Node != "" && element.Service != "" {
			name := element.Node
			noodles[name] = make(chan gotocol.Message)
			// start the service and tell it it's name
			switch element.Service {
			case "pirate":
				go pirate.Start(noodles[name])
				noodles[name] <- gotocol.Message{gotocol.Hello, listener, name}
				if edda.Logchan != nil {
					// tell the pirate to report itself and new edges to the logger
					noodles[name] <- gotocol.Message{gotocol.Inform, edda.Logchan, ""}
				}
			default:
				log.Println("fsm: unknown service: " + element.Service)
			}
		}
	}
	// Make all the connections
	for _, element := range g.Graph {
		if element.Edge != "" && element.Source != "" && element.Target != "" {
			noodles[element.Source] <- gotocol.Message{gotocol.NameDrop, noodles[element.Target], element.Target}
			log.Println("Link " + element.Source + " > " + element.Target)
		}
	}
	// send money and start the pirates chatting
	for _, noodle := range noodles {
		// same as below for now, but will save and read back from file later
		// anonymously send this pirate a random amount of GoldCoin up to 100
		gold := fmt.Sprintf("%d", rand.Intn(100))
		noodle <- gotocol.Message{gotocol.GoldCoin, nil, gold}
		// tell this pirate to start chatting with friends every 0.1 to 10 secs
		delay := fmt.Sprintf("%dms", 100+rand.Intn(9900))
		noodle <- gotocol.Message{gotocol.Chat, nil, delay}
	}
	shutdown()
}

// Start fsm and create new pirates
func Start() {
	pirate.Msglog = Msglog                // pass on console message log flag if set
	listener = make(chan gotocol.Message) // listener for fsm
	if Population < 2 {
		log.Fatal("fsm: can't create less than 2 pirates")
	}
	// create map of channels and a name index to select randoml nodes from
	noodles = make(map[string]chan gotocol.Message, Population)
	names = make([]string, Population) // indexable name list
	log.Println("fsm: population", Population, "pirates")
	for i := 1; i <= Population; i++ {
		name := fmt.Sprintf("Pirate%d", i)
		noodles[name] = make(chan gotocol.Message)
		go pirate.Start(noodles[name])
	}
	i := 0
	msgcount := 1
	start := time.Now()
	for name, noodle := range noodles {
		names[i] = name
		i++
		// tell the pirate it's name and how to talk back to it's fsm
		// this must be the first message the pirate sees
		noodle <- gotocol.Message{gotocol.Hello, listener, name}
		if edda.Logchan != nil {
			// tell the pirate to report itself and new edges to the logger
			noodle <- gotocol.Message{gotocol.Inform, edda.Logchan, ""}
			msgcount = 2
		}
	}
	log.Println("fsm: Talk amongst yourselves for", ChatSleep)
	rand.Seed(int64(len(noodles)))
	for _, name := range names {
		// for each pirate tell them about two other random pirates
		noodle := noodles[name] // lookup the channel
		// pick a first random pirate to tell this one about
		talkto := names[rand.Intn(len(names))]
		noodle <- gotocol.Message{gotocol.NameDrop, noodles[talkto], talkto}
		// pick a second random pirate to tell this one about
		talkto = names[rand.Intn(len(names))]
		noodle <- gotocol.Message{gotocol.NameDrop, noodles[talkto], talkto}
		// anonymously send this pirate a random amount of GoldCoin up to 100
		gold := fmt.Sprintf("%d", rand.Intn(100))
		noodle <- gotocol.Message{gotocol.GoldCoin, nil, gold}
		// tell this pirate to start chatting with friends every 0.1 to 10 secs
		delay := fmt.Sprintf("%dms", 100+rand.Intn(9900))
		noodle <- gotocol.Message{gotocol.Chat, nil, delay}
	}
	msgcount += 4
	d := time.Since(start)
	log.Println("fsm: Delivered", msgcount*len(names), "messages in", d)
	shutdown()
}

// Shutdown fsm and pirates
func shutdown() {
	var msg gotocol.Message
	// wait until the delay has finished
	if ChatSleep >= time.Millisecond {
		time.Sleep(ChatSleep)
	}
	log.Println("fsm: Shutdown")
	for _, noodle := range noodles {
		gotocol.Message{gotocol.Goodbye, nil, "beer volcano"}.GoSend(noodle)
	}
	for len(noodles) > 0 {
		msg = <-listener
		if Msglog {
			log.Printf("fsm: %v\n", msg)
		}
		switch msg.Imposition {
		case gotocol.Goodbye:
			delete(noodles, msg.Intention)
			if Msglog {
				log.Printf("fsm: Pirate population: %v    \n", len(noodles))
			}
		}
	}
	if edda.Logchan != nil {
		close(edda.Logchan)
	}
	log.Println("fsm: Exit")
}
