// Package fsm implements a Flexible State Manager (Flying Spaghetti Monster)
// It creates and controls a large social network of pirates via channels (the noodly touch)
// and logs the architecture (nodes and links) as it evolves
package fsm

import (
	"fmt"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/logger"
	"log"
	"math/rand"
	"time"
)

// ChatSleep duration is set via command line flag to tell fsm how long to let pirates chat
var ChatSleep time.Duration

// Msglog toggles whether to log every message received to the consolei
var Msglog bool

// Touch all the noodles that connect to the pirates to manage them
func Touch(noodles map[string]chan gotocol.Message) {
	var msg gotocol.Message
	names := make([]string, len(noodles))  // indexable name list
	listener := make(chan gotocol.Message) // listener for fsm
	log.Println("fsm: Hello")
	i := 0
	msgcount := 1
	start := time.Now()
	for name, noodle := range noodles {
		names[i] = name
		i++
		// tell the pirate it's name and how to talk back to it's fsm
		// this must be the first message the pirate sees
		noodle <- gotocol.Message{gotocol.Hello, listener, name}
		if logger.Logchan != nil {
			// tell the pirate to report itself and new edges to the logger
			noodle <- gotocol.Message{gotocol.Inform, logger.Logchan, ""}
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
	if ChatSleep >= time.Millisecond {
		time.Sleep(ChatSleep)
	}
	log.Println("fsm: Go away")
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
	if logger.Logchan != nil {
		close(logger.Logchan)
	}
	log.Println("fsm: Exit")
}
