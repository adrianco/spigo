// Flexible State Manager (a.k.a. Flying Spaghetti Monster)
// Controls a large collection of pirates, touching with its noodles
// Logs the architecture (nodes and links) as it evolves

package fsm

import (
	"fmt"
	"log"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/graphml"
	"github.com/adrianco/spigo/logger"
	"math/rand"
	"time"
)

var ChatSleep time.Duration

// FSM touches all the noodles that connect to the pirates etc.
func Touch(noodles map[string]chan gotocol.Message) {
	var msg gotocol.Message
	names := make([]string, len(noodles))  // indexable name list
	listener := make(chan gotocol.Message) // listener for fsm
	log.Println("fsm: Hello")
	i := 0
	msgcount := 0
	for name, noodle := range noodles {
		graphml.WriteNode(name)
		graphjson.WriteNode(name, "pirate")
		noodle <- gotocol.Message{gotocol.Hello, listener, name}
		names[i] = name
		i = i + 1
		if logger.Logchan != nil {
			// tell the pirate to report new edges to the logger
			noodle <- gotocol.Message{gotocol.Inform, logger.Logchan, ""}
			msgcount = 1
		}
	}
	log.Println("fsm: Talk amongst yourselves for", ChatSleep)
	rand.Seed(int64(len(noodles)))
	start := time.Now()
	for i := 0; i < len(names); i++ {
		// for each pirate tell them about two other random pirates
		noodle := noodles[names[i]] // lookup the channel
		// pick a first random pirate to tell this one about
		talkto := names[rand.Intn(len(names))]
		noodle <- gotocol.Message{gotocol.NameDrop, noodles[talkto], talkto}
		// pick a second random pirate to tell this one about
		talkto = names[rand.Intn(len(names))]
		noodle <- gotocol.Message{gotocol.NameDrop, noodles[talkto], talkto}
		// send this pirate a random amount of GoldCoin up to 100
		gold := fmt.Sprintf("%d", rand.Intn(100))
		noodle <- gotocol.Message{gotocol.GoldCoin, listener, gold}
		// tell this pirate to start chatting with friends every 1-60s
		delay := fmt.Sprintf("%ds", 1+rand.Intn(59))
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
		// fmt.Printf("fsm: %v\n", msg)
		switch msg.Imposition {
		case gotocol.Goodbye:
			delete(noodles, msg.Intention)
			fmt.Printf("fsm: Pirate population: %v    \r", len(noodles))
		}
	}
	if logger.Logchan != nil {
		close(logger.Logchan)
	}
	log.Println("fsm: Exit")
}
