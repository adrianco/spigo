// Package monolith simulates a monolithic business logic microservice
// Takes incoming traffic and calls into dependent microservices across all zones
package monolith

import (
	"github.com/adrianco/spigo/tooling/archaius"
	"github.com/adrianco/spigo/tooling/collect"
	"github.com/adrianco/spigo/tooling/flow"
	"github.com/adrianco/spigo/tooling/gotocol"
	"github.com/adrianco/spigo/tooling/handlers"
	"github.com/adrianco/spigo/tooling/ribbon"
	"time"
)

// Start - all configuration and state is sent via messages
func Start(listener chan gotocol.Message) {
	microservices := ribbon.MakeRouter()
	dependencies := make(map[string]time.Time)         // dependent services and time last updated
	var parent chan gotocol.Message                    // remember how to talk back to creator
	requestor := make(map[string]gotocol.Routetype)    // remember where requests came from when responding
	var name string                                    // remember my name
	eureka := make(map[string]chan gotocol.Message, 1) // service registry
	hist := collect.NewHist("")
	ep, _ := time.ParseDuration(archaius.Conf.EurekaPoll)
	eurekaTicker := time.NewTicker(ep)
	for {
		select {
		case msg := <-listener:
			flow.Instrument(msg, name, hist)
			switch msg.Imposition {
			case gotocol.Hello:
				if name == "" {
					// if I don't have a name yet remember what I've been named
					parent = msg.ResponseChan // remember how to talk to my namer
					name = msg.Intention      // message body is my name
					hist = collect.NewHist(name)
				}
			case gotocol.Inform:
				eureka[msg.Intention] = handlers.Inform(msg, name, listener)
			case gotocol.NameDrop:
				handlers.NameDrop(&dependencies, microservices, msg, name, listener, eureka, true) // true to setup cross region routing
			case gotocol.Forget:
				// forget a buddy
				handlers.Forget(&dependencies, microservices, msg)
			case gotocol.GetRequest:
				// route the request on to microservices
				handlers.GetRequest(msg, name, listener, &requestor, microservices)
			case gotocol.GetResponse:
				// return path from a request, send payload back up using saved span context - server send
				handlers.GetResponse(msg, name, listener, &requestor)
			case gotocol.Put:
				// route the request on to a random dependency
				handlers.Put(msg, name, listener, &requestor, microservices)
			case gotocol.Goodbye:
				for _, ch := range eureka { // tell name service I'm not going to be here
					ch <- gotocol.Message{gotocol.Delete, nil, time.Now(), gotocol.NilContext, name}
				}
				gotocol.Message{gotocol.Goodbye, nil, time.Now(), gotocol.NilContext, name}.GoSend(parent)
				return
			}
		case <-eurekaTicker.C: // check to see if any new dependencies have appeared
			for dep, _ := range dependencies {
				for _, ch := range eureka {
					ch <- gotocol.Message{gotocol.GetRequest, listener, time.Now(), gotocol.NilContext, dep}
				}
			}
		}
	}
}
