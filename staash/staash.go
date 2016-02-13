// Package staash simulates a generic data access layer microservice
// Takes incoming traffic and calls into dependent microservices in a single zone
package staash

import (
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/collect"
	"github.com/adrianco/spigo/flow"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/handlers"
	. "github.com/adrianco/spigo/packagenames"
	"github.com/adrianco/spigo/ribbon"
	"time"
)

// Start staash, all configuration and state is sent via messages
func Start(listener chan gotocol.Message) {
	microservices := ribbon.MakeRouter()               // outbound routes
	var caches, stores, volumes, cass *ribbon.Router   // subsets of the router
	dependencies := make(map[string]time.Time)         // dependent service names and time last updated
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
				handlers.NameDrop(&dependencies, microservices, msg, name, listener, eureka, true) // true to setup cross zone routing
				caches = microservices.All(CachePkg)
				volumes = microservices.All(VolumePkg)
				stores = microservices.All(StorePkg)
				cass = microservices.All(PriamCassandraPkg)
			case gotocol.Forget:
				// forget a buddy
				handlers.Forget(&dependencies, microservices, msg)
				caches = microservices.All(CachePkg)
				volumes = microservices.All(VolumePkg)
				stores = microservices.All(StorePkg)
				cass = microservices.All(PriamCassandraPkg)
			case gotocol.GetRequest:
				// route the request on to a cache first if configured
				if caches.Len() > 0 {
					handlers.GetRequest(msg, name, listener, &requestor, caches)
				} else {
					// route to any volumes next if configured
					if volumes.Len() > 0 {
						handlers.GetRequest(msg, name, listener, &requestor, volumes)
					} else {
						// route to any cassandra if configured
						if cass.Len() > 0 {
							handlers.GetRequest(msg, name, listener, &requestor, cass)
						} else {
							// route to stores if configured
							if stores.Len() > 0 {
								handlers.GetRequest(msg, name, listener, &requestor, stores)
							}
						}
					}
				}
			case gotocol.GetResponse:
				// return path from a request, send payload back up using saved span context - server send
				handlers.GetResponse(msg, name, listener, &requestor)
			case gotocol.Put:
				// duplicate the request to any cache, volumes, stores, and cassandra but only to one of each type
				// storage class packages sideways Replicate if configured
				// to get a lossy write, configure multiple stores that don't cross replicate
				handlers.Put(msg, name, listener, &requestor, caches)
				handlers.Put(msg, name, listener, &requestor, volumes)
				handlers.Put(msg, name, listener, &requestor, stores)
				handlers.Put(msg, name, listener, &requestor, cass)
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
