// Package fsm implements a Flexible State Manager (Flying Spaghetti Monster)
// It creates and controls a large social network of pirates via channels (the noodly touch)
// or reads in a network from a json file. I also logs the architecture (nodes and links) as it evolves
package fsm

import (
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/collect"
	"github.com/adrianco/spigo/edda"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/names"
	"github.com/adrianco/spigo/pirate"
	"log"
	"math/rand"
	"time"
)

// noodles channels mapped by pirate name connects fsm to everyone
var noodles map[string]chan gotocol.Message
var pnames []string
var listener chan gotocol.Message

// Reload the network from a file
func Reload(arch string) {
	listener = make(chan gotocol.Message) // listener for fsm
	log.Println("fsm reloading from " + arch + ".json")
	g := graphjson.ReadArch(arch)
	pop := 0
	// count how many nodes there are
	for _, element := range g.Graph {
		if element.Node != "" {
			pop++
		}
	}
	archaius.Conf.Population = pop
	// create the map of channels
	noodles = make(map[string]chan gotocol.Message, archaius.Conf.Population)
	// Start all the services
	for _, element := range g.Graph {
		if element.Node != "" && element.Service != "" {
			name := element.Node
			noodles[name] = make(chan gotocol.Message)
			// start the service and tell it it's name
			switch element.Service {
			case "pirate":
				go pirate.Start(noodles[name])
				noodles[name] <- gotocol.Message{gotocol.Hello, listener, time.Now(), gotocol.NilContext(), name}
				if edda.Logchan != nil {
					// tell the pirate to report itself and new edges to the logger
					noodles[name] <- gotocol.Message{gotocol.Inform, edda.Logchan, time.Now(), gotocol.NilContext(), ""}
				}
			default:
				log.Println("fsm: unknown service: " + element.Service)
			}
		}
	}
	// Make all the connections
	for _, element := range g.Graph {
		if element.Edge != "" && element.Source != "" && element.Target != "" {
			noodles[element.Source] <- gotocol.Message{gotocol.NameDrop, noodles[element.Target], time.Now(), gotocol.NewRequest(), element.Target}
			log.Println("Link " + element.Source + " > " + element.Target)
		}
	}
	// send money and start the pirates chatting
	for _, noodle := range noodles {
		// same as below for now, but will save and read back from file later
		// anonymously send this pirate a random amount of GoldCoin up to 100
		gold := fmt.Sprintf("%d", rand.Intn(100))
		noodle <- gotocol.Message{gotocol.GoldCoin, nil, time.Now(), gotocol.NewRequest(), gold}
		// tell this pirate to start chatting with friends every 0.1 to 10 secs
		delay := fmt.Sprintf("%dms", 100+rand.Intn(9900))
		noodle <- gotocol.Message{gotocol.Chat, nil, time.Now(), gotocol.NilContext(), delay}
	}
	shutdown()
}

// Start fsm and create new pirates
func Start() {
	listener = make(chan gotocol.Message) // listener for fsm
	if archaius.Conf.Population < 2 {
		log.Fatal("fsm: can't create less than 2 pirates")
	}
	// create map of channels and a name index to select randoml nodes from
	noodles = make(map[string]chan gotocol.Message, archaius.Conf.Population)
	pnames = make([]string, archaius.Conf.Population) // indexable name list
	log.Println("fsm: population", archaius.Conf.Population, "pirates")
	for i := 1; i <= archaius.Conf.Population; i++ {
		name := names.Make(archaius.Conf.Arch, "atlantic", "bermuda", "blackbeard", "pirate", i)
		noodles[name] = make(chan gotocol.Message)
		go pirate.Start(noodles[name])
	}
	i := 0
	msgcount := 1
	start := time.Now()
	for name, noodle := range noodles {
		pnames[i] = name
		i++
		// tell the pirate it's name and how to talk back to it's fsm
		// this must be the first message the pirate sees
		noodle <- gotocol.Message{gotocol.Hello, listener, time.Now(), gotocol.NilContext(), name}
		if edda.Logchan != nil {
			// tell the pirate to report itself and new edges to the logger
			noodle <- gotocol.Message{gotocol.Inform, edda.Logchan, time.Now(), gotocol.NilContext(), ""}
			msgcount = 2
		}
	}
	log.Println("fsm: Talk amongst yourselves for", archaius.Conf.RunDuration)
	rand.Seed(int64(len(noodles)))
	for _, name := range pnames {
		// for each pirate tell them about two other random pirates
		noodle := noodles[name] // lookup the channel
		// pick a first random pirate to tell this one about
		talkto := pnames[rand.Intn(len(pnames))]
		noodle <- gotocol.Message{gotocol.NameDrop, noodles[talkto], time.Now(), gotocol.NewRequest(), talkto}
		// pick a second random pirate to tell this one about
		talkto = pnames[rand.Intn(len(pnames))]
		noodle <- gotocol.Message{gotocol.NameDrop, noodles[talkto], time.Now(), gotocol.NewRequest(), talkto}
		// anonymously send this pirate a random amount of GoldCoin up to 100
		gold := fmt.Sprintf("%d", rand.Intn(100))
		noodle <- gotocol.Message{gotocol.GoldCoin, nil, time.Now(), gotocol.NewRequest(), gold}
		// tell this pirate to start chatting with friends every 0.1 to 10 secs
		delay := fmt.Sprintf("%dms", 100+rand.Intn(9900))
		noodle <- gotocol.Message{gotocol.Chat, nil, time.Now(), gotocol.NewRequest(), delay}
	}
	msgcount += 4
	d := time.Since(start)
	log.Println("fsm: Delivered", msgcount*len(pnames), "messages in", d)
	shutdown()
}

// Shutdown fsm and pirates
func shutdown() {
	var msg gotocol.Message
	hist := collect.NewHist("fsm")
	// wait until the delay has finished
	if archaius.Conf.RunDuration >= time.Millisecond {
		time.Sleep(archaius.Conf.RunDuration)
	}
	log.Println("fsm: Shutdown")
	for _, noodle := range noodles {
		gotocol.Message{gotocol.Goodbye, nil, time.Now(), gotocol.NilContext(), "beer volcano"}.GoSend(noodle)
	}
	for len(noodles) > 0 {
		msg = <-listener
		collect.Measure(hist, time.Since(msg.Sent))
		if archaius.Conf.Msglog {
			log.Printf("fsm: %v\n", msg)
		}
		switch msg.Imposition {
		case gotocol.Goodbye:
			delete(noodles, msg.Intention)
			if archaius.Conf.Msglog {
				log.Printf("fsm: Pirate population: %v    \n", len(noodles))
			}
		}
	}
	collect.Save()
	log.Println("fsm: Exit")
}
