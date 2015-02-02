// Package elb simulates an Elastic Load Balancer
// Takes incoming traffic and spreads it over microservices in three availability zones
package elb

import (
	"github.com/adrianco/spigo/gotocol"
	"log"
	"math/rand"
	"time"
)

// Msglog turns on console logging of messages
var Msglog bool

// Start the elb, all configuration and state is sent via messages
func Start(listener chan gotocol.Message) {
	dunbar := 30 // starting point for how many nodes to remember
	// remember the channel to talk to microservices
	microservices := make(map[string]chan gotocol.Message, dunbar)
	microindex := make([]chan gotocol.Message, dunbar)
	var netflixoss chan gotocol.Message // remember how to talk back to creator
	var name string                     // remember my name
	var edda chan gotocol.Message       // if set, send updates
	var chatrate time.Duration
	chatTicker := time.NewTicker(time.Hour)
	chatTicker.Stop()
	for {
		select {
		case msg := <-listener:
			if Msglog {
				log.Printf("%v: %v\n", name, msg)
			}
			switch msg.Imposition {
			case gotocol.Hello:
				if name == "" {
					// if I don't have a name yet remember what I've been named
					netflixoss = msg.ResponseChan // remember how to talk to my namer
					name = msg.Intention          // message body is my name
				}
			case gotocol.Inform:
				// remember where to send updates
				edda = msg.ResponseChan
				// logger channel is buffered so no need to use GoSend
				edda <- gotocol.Message{gotocol.Hello, nil, name + " " + "elb"}
			case gotocol.NameDrop:
				// don't remember too many buddies and don't talk to myself
				microservice := msg.Intention // message body is buddy name
				if len(microservices) < dunbar && microservice != name {
					// remember how to talk to this buddy
					microservices[microservice] = msg.ResponseChan // message channel is buddy's listener
					if edda != nil {
						// if it's setup, tell the logger I have a new buddy to talk to
						edda <- gotocol.Message{gotocol.Inform, listener, name + " " + microservice}
					}
				}
			case gotocol.Chat:
				// setup the ticker to run at the specified rate
				d, e := time.ParseDuration(msg.Intention)
				if e == nil && d >= time.Millisecond && d <= time.Hour {
					chatrate = d
					chatTicker = time.NewTicker(chatrate)
				}
			case gotocol.GetResponse:
				// return path from a request
				// nothing to do at this level
			case gotocol.Goodbye:
				if Msglog {
					log.Printf("%v: Going away, chatting every %v\n", name, chatrate)
				}
				gotocol.Message{gotocol.Goodbye, nil, name}.GoSend(netflixoss)
				return
			}
		case <-chatTicker.C:
			if len(microservices) > 0 {
				// build index if needed
				if len(microindex) != len(microservices) {
					i := 0
					for _, ch := range microservices {
						microindex[i] = ch
						i++
					}
				}
				m := rand.Intn(len(microservices))
				// start a request to a random member of this elb
				gotocol.Message{gotocol.GetRequest, listener, name}.GoSend(microindex[m])
			}
			//default:
		}
	}
}
