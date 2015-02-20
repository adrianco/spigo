// Package staash simulates a data access layer microservice
// Takes incoming traffic and calls into cassandra  microservices in a single zone
// Code is a pure clone of Karyon to start with
package staash

import (
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/collect"
	"github.com/adrianco/spigo/gotocol"
	"log"
	"math/rand"
	"time"
)

// Start staash, all configuration and state is sent via messages
func Start(listener chan gotocol.Message) {
	dunbar := 30 // starting point for how many nodes to remember
	// remember the channel to talk to microservices
	microservices := make(map[string]chan gotocol.Message, dunbar)
	microindex := make([]chan gotocol.Message, dunbar)
	store := make(map[string]string, 4)            // key value store
	var netflixoss, requestor chan gotocol.Message // remember creator and how to talk back to incoming requests
	var name string                                // remember my name
	var edda chan gotocol.Message                  // if set, send updates
	var chatrate time.Duration
	hist := collect.NewHist("")
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
				edda <- gotocol.Message{gotocol.Hello, nil, time.Now(), name + " " + "staash"}
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
			case gotocol.GetRequest:
				// route the request on to microservices
				requestor = msg.ResponseChan
				// Intention body indicates which service to route to or which key to get
				// need to lookup service by type rather than randomly call one day
				if len(microservices) > 0 {
					if len(microindex) != len(microservices) {
						// rebuild index
						i := 0
						for _, ch := range microservices {
							microindex[i] = ch
							i++
						}
					}
					m := rand.Intn(len(microservices))
					// start a request to a random microservice
					gotocol.Message{gotocol.GetRequest, listener, time.Now(), msg.Intention}.GoSend(microindex[m])
				}
			case gotocol.GetResponse:
				// return path from a request, send payload back up
				if requestor != nil {
					gotocol.Message{gotocol.GetResponse, listener, time.Now(), msg.Intention}.GoSend(requestor)
				}
			case gotocol.Put:
				// route the request on to a random dependency
				if len(microservices) > 0 {
					if len(microindex) != len(microservices) {
						// rebuild index
						i := 0
						for _, ch := range microservices {
							microindex[i] = ch
							i++
						}
					}
					m := rand.Intn(len(microservices))
					// pass on request to a random service
					gotocol.Message{gotocol.Put, listener, time.Now(), msg.Intention}.GoSend(microindex[m])
				}
			case gotocol.Goodbye:
				if archaius.Conf.Msglog {
					log.Printf("%v: Going away, zone: %v\n", name, store["zone"])
				}
				gotocol.Message{gotocol.Goodbye, nil, time.Now(), name}.GoSend(netflixoss)
				return
			}
		case <-chatTicker.C:
			if len(microservices) > 0 {
				if len(microservices) != len(microindex) {
					// rebuild index
					i := 0
					for _, ch := range microservices {
						microindex[i] = ch
						i++
					}
				}
				m := rand.Intn(len(microservices))
				// start a request to a random member of this elb
				gotocol.Message{gotocol.GetRequest, listener, time.Now(), name}.GoSend(microindex[m])
			}
			//default:
		}
	}
}
