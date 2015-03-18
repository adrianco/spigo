// Package zuul simulates a generic business logic microservice
// Takes incoming traffic and calls into dependent microservices in a single zone
package zuul

import (
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/collect"
	"github.com/adrianco/spigo/gotocol"
	"log"
	"math/rand"
	"time"
)

// Start zuul, all configuration and state is sent via messages
func Start(listener chan gotocol.Message) {
	dunbar := 30 // starting point for how many nodes to remember
	// remember the channel to talk to microservices
	microservices := make(map[string]chan gotocol.Message, dunbar)
	microindex := make([]chan gotocol.Message, dunbar)
	dependencies := make(map[string]time.Time, dunbar) // dependent services and time last updated
	var netflixoss, requestor chan gotocol.Message // remember creator and how to talk back to incoming requests
	var name string                                // remember my name
	eureka := make(map[string]chan gotocol.Message, 1) // service registry
	var chatrate time.Duration
	hist := collect.NewHist("")
	ep, _ := time.ParseDuration(archaius.Conf.EurekaPoll)
	eurekaTicker := time.NewTicker(ep)
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
				eureka[msg.Intention] = gotocol.InformHandler(msg, name, listener)
			case gotocol.NameDrop:
				gotocol.NameDropHandler(&dependencies, &microservices, msg, name, listener, eureka)
			case gotocol.Forget:
				// forget a buddy
				delete(microservices, msg.Intention)
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
					// start a request to a random service
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
					log.Printf("%v: Going away\n", name)
				}
				gotocol.Message{gotocol.Goodbye, nil, time.Now(), name}.GoSend(netflixoss)
				return
			}
		case <-eurekaTicker.C: // check to see if any new dependencies have appeared
			for dep, _ := range dependencies {
				for _, ch := range eureka {
					ch <- gotocol.Message{gotocol.GetRequest, listener, time.Now(), dep}
				}
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
