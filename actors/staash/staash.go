// Package staash simulates a generic data access layer microservice
// Takes incoming traffic and calls into dependent microservices in a single zone
package staash

import (
	. "github.com/adrianco/spigo/actors/packagenames"
	"github.com/adrianco/spigo/tooling/archaius"
	"github.com/adrianco/spigo/tooling/collect"
	"github.com/adrianco/spigo/tooling/flow"
	"github.com/adrianco/spigo/tooling/gotocol"
	"github.com/adrianco/spigo/tooling/handlers"
	"github.com/adrianco/spigo/tooling/names"
	"github.com/adrianco/spigo/tooling/ribbon"
	"time"
)

// states for request resolution stored in gotocol.Routetype.State
const (
	newRequest = iota
	cacheLookup
	volumeLookup
	cassandraLookup
	storeLookup
	staashLookup
)

// Start staash, all configuration and state is sent via messages
func Start(listener chan gotocol.Message) {
	microservices := ribbon.MakeRouter()                     // outbound routes
	var caches, stores, volumes, cass, staash *ribbon.Router // subsets of the router
	dependencies := make(map[string]time.Time)               // dependent service names and time last updated
	var parent chan gotocol.Message                          // remember how to talk back to creator
	requestor := make(map[string]gotocol.Routetype)          // remember where requests came from when responding
	var name string                                          // remember my name
	eureka := make(map[string]chan gotocol.Message, 1)       // service registry
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
				staash = microservices.All(StaashPkg)
			case gotocol.Forget:
				// forget a buddy
				handlers.Forget(&dependencies, microservices, msg)
				caches = microservices.All(CachePkg)
				volumes = microservices.All(VolumePkg)
				stores = microservices.All(StorePkg)
				cass = microservices.All(PriamCassandraPkg)
				staash = microservices.All(StaashPkg)
			case gotocol.GetRequest:
				// route the request on to a cache first if configured
				r := gotocol.PickRoute(requestor, msg)
				if caches.Len() > 0 {
					handlers.GetRequest(msg, name, listener, &requestor, caches)
					r.State = cacheLookup
				} else {
					// route to any volumes if configured
					if volumes.Len() > 0 {
						handlers.GetRequest(msg, name, listener, &requestor, volumes)
						r.State = volumeLookup
					} else {
						// route to any cassandra if configured
						if cass.Len() > 0 {
							handlers.GetRequest(msg, name, listener, &requestor, cass)
							r.State = cassandraLookup
						} else {
							// route to stores if configured
							if stores.Len() > 0 {
								handlers.GetRequest(msg, name, listener, &requestor, stores)
								r.State = storeLookup
							} else {
								// route to more staash layers if configured
								if staash.Len() > 0 {
									handlers.GetRequest(msg, name, listener, &requestor, staash)
									r.State = staashLookup
								}
							}
						}
					}
				}
			case gotocol.GetResponse:
				// return path from a request, resend or send payload back up using saved span context - server send
				r := gotocol.PickRoute(requestor, msg)
				if msg.Intention != "" { // we got a value, so pass it back up
					handlers.GetResponse(msg, name, listener, &requestor)
				} else {
					switch r.State {
					case cacheLookup:
						if volumes.Len() > 0 {
							handlers.GetRequest(msg, name, listener, &requestor, volumes)
							r.State = volumeLookup
							break
						}
						fallthrough // no volumes so look for cassandra
					case volumeLookup:
						if cass.Len() > 0 {
							handlers.GetRequest(msg, name, listener, &requestor, cass)
							r.State = cassandraLookup
							break
						}
						fallthrough // no cassandra so look for stores
					case cassandraLookup:
						if stores.Len() > 0 {
							handlers.GetRequest(msg, name, listener, &requestor, stores)
							r.State = storeLookup
							break
						}
						fallthrough // no stores
					case storeLookup:
						if staash.Len() > 0 {
							handlers.GetRequest(msg, name, listener, &requestor, staash)
							r.State = staashLookup
							break
						}
						fallthrough // no staash
					case staashLookup:
						// ran out of options to find anything so pass empty response back up
						handlers.GetResponse(msg, name, listener, &requestor)
					case newRequest:
					default:
					}
				}
			case gotocol.Put:
				// duplicate the request to any cache, volumes, stores, and cassandra but only to one of each type
				// storage class packages sideways Replicate if configured
				// to get a lossy write, configure multiple stores that don't cross replicate
				handlers.Put(msg, name, listener, &requestor, caches)
				handlers.Put(msg, name, listener, &requestor, cass)
				handlers.Put(msg, name, listener, &requestor, staash)
				handlers.Put(msg, name, listener, &requestor, stores)
				msg.Intention = names.Instance(name) + "/" + msg.Intention // store to an instance specific volume namespace
				handlers.Put(msg, name, listener, &requestor, volumes)
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
