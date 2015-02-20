// Package denominator simulates a global DNS service
// Takes incoming traffic and spreads it over elb's in multiple regions
package denominator

import (
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/collect"
	"github.com/adrianco/spigo/gotocol"
	"log"
	"math/rand"
	"time"
)

// Start the denominator, all configuration and state is sent via messages
func Start(listener chan gotocol.Message) {
	dunbar := 5 // starting point for how many nodes to remember
	// remember the channel to talk to microservices
	microservices := make(map[string]chan gotocol.Message, dunbar)
	microindex := make([]chan gotocol.Message, dunbar)
	var netflixoss chan gotocol.Message // remember how to talk back to creator
	var name string                     // remember my name
	hist := collect.NewHist("")         // don't know name yet
	var edda chan gotocol.Message       // if set, send updates
	var chatrate time.Duration
	chatTicker := time.NewTicker(time.Hour)
	chatTicker.Stop()
	for {
		select {
		case msg := <-listener:
			collect.Measure(hist, time.Since(msg.Sent))
			if archaius.Conf.Msglog {
				log.Printf("%v: %v\n", name, msg)
			}
			switch msg.Imposition {
			case gotocol.Hello:
				if name == "" {
					// if I don't have a name yet remember what I've been named
					netflixoss = msg.ResponseChan // remember how to talk to my namer
					name = msg.Intention          // message body is my name
					hist = collect.NewHist(name)
				}
			case gotocol.Inform:
				// remember where to send updates
				edda = msg.ResponseChan
				// logger channel is buffered so no need to use GoSend
				edda <- gotocol.Message{gotocol.Hello, nil, time.Now(), name + " " + "denominator"}
			case gotocol.NameDrop:
				// don't remember too many buddies and don't talk to myself
				microservice := msg.Intention // message body is buddy name
				if len(microservices) < dunbar && microservice != name {
					// remember how to talk to this buddy
					microservices[microservice] = msg.ResponseChan // message channel is buddy's listener
					if edda != nil {
						// if it's setup, tell the logger I have a new buddy to talk to
						edda <- gotocol.Message{gotocol.Inform, listener, time.Now(), name + " " + microservice}
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
				if archaius.Conf.Msglog {
					log.Printf("%v: Going away, was chatting every %v\n", name, chatrate)
				}
				gotocol.Message{gotocol.Goodbye, nil, time.Now(), name}.GoSend(netflixoss)
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
				// start a request to a random member of this denominator
				if rand.Intn(2) == 0 {
					gotocol.Message{gotocol.GetRequest, listener, time.Now(), "why?"}.GoSend(microindex[m])
				} else {
					gotocol.Message{gotocol.Put, listener, time.Now(), "remember me"}.GoSend(microindex[m])
				}
			}
		}
	}
}
