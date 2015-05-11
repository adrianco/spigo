// Package priamCassandra simulates a generic business logic microservice
// Takes incoming traffic and calls into dependent microservices in a single zone
package priamCassandra

import (
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/collect"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/names"
	"log"
	"time"
)

// Start priamCassandra, all configuration and state is sent via messages
func Start(listener chan gotocol.Message) {
	dunbar := 100 // starting point for how many nodes to remember
	// remember the channel to talk to microservices
	microservices := make(map[string]chan gotocol.Message, dunbar)
	dependencies := make(map[string]time.Time, dunbar) // dependent services and time last updated
	store := make(map[string]string, 4)                // key value store
	store["why?"] = "because..."
	var netflixoss, requestor chan gotocol.Message                           // remember creator and how to talk back to incoming requests
	var name string                                                          // remember my name
	eureka := make(map[string]chan gotocol.Message, 3*archaius.Conf.Regions) // service registry per zone and region
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
			case gotocol.NameDrop: // cross zone = true
				gotocol.NameDropHandler(&dependencies, &microservices, msg, name, listener, eureka, true)
			case gotocol.Forget:
				// forget a buddy
				gotocol.ForgetHandler(&dependencies, &microservices, msg)
			case gotocol.Chat:
				// setup the ticker to run at the specified rate
				d, e := time.ParseDuration(msg.Intention)
				if e == nil && d >= time.Millisecond && d <= time.Hour {
					chatrate = d
					chatTicker = time.NewTicker(chatrate)
				}
			case gotocol.GetRequest:
				// return any stored value for this key (Cassandra READ.ONE behavior)
				gotocol.Message{gotocol.GetResponse, listener, time.Now(), store[msg.Intention]}.GoSend(msg.ResponseChan)
			case gotocol.GetResponse:
				// return path from a request, send payload back up (not currently used)
				if requestor != nil {
					gotocol.Message{gotocol.GetResponse, listener, time.Now(), msg.Intention}.GoSend(requestor)
				}
			case gotocol.Put:
				requestor = msg.ResponseChan
				// set a key value pair and replicate globally
				var key, value string
				fmt.Sscanf(msg.Intention, "%s%s", &key, &value)
				if key != "" && value != "" {
					store[key] = value
					// duplicate the request on to priamCassandra nodes in each zone and one in each region
					for _, z := range names.OtherZones(name, archaius.Conf.ZoneNames) {
						// replicate request
						for n, c := range microservices {
							if names.Region(n) == names.Region(name) && names.Zone(n) == z {
								gotocol.Message{gotocol.Replicate, listener, time.Now(), msg.Intention}.GoSend(c)
								break // only need to send it to one node in each zone, no tokens yet
							}
						}
					}
					for _, r := range names.OtherRegions(name, archaius.Conf.RegionNames[0:archaius.Conf.Regions]) {
						for n, c := range microservices {
							if names.Region(n) == r {
								gotocol.Message{gotocol.Replicate, listener, time.Now(), msg.Intention}.GoSend(c)
								break // only need to send it to one node in each region, no tokens yet
							}
						}
					}
				}
			case gotocol.Replicate:
				// Replicate is only used between priamCassandra nodes
				// end point for a request
				var key, value string
				fmt.Sscanf(msg.Intention, "%s%s", &key, &value)
				// log.Printf("priamCassandra: %v:%v", key, value)
				if key != "" && value != "" {
					store[key] = value
				}
				// name looks like: netflixoss.us-east-1.zoneC.cassTurtle.priamCassandra.cassTurtle11
				myregion := names.Region(name)
				//log.Printf("%v: %v\n", name, myregion)
				// find if this was a cross region Replicate
				for in, c := range microservices {
					// find the name matching incoming request channel to see where its coming from
					if c == msg.ResponseChan && myregion != names.Region(in) {
						// Replicate from out of region needs to be Replicated once only to other zones in this Region
						for _, z := range names.OtherZones(name, archaius.Conf.ZoneNames) {
							// replicate request
							for n, c := range microservices {
								if names.Region(n) == myregion && names.Zone(n) == z {
									gotocol.Message{gotocol.Replicate, listener, time.Now(), msg.Intention}.GoSend(c)
									break // only need to send it to one node in each zone, no tokens yet
								}
							}
						}
						break
					}
				}
			case gotocol.Goodbye:
				if archaius.Conf.Msglog {
					log.Printf("%v: Going away, zone: %v\n", name, store["zone"])
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
			// nothing to do here at the moment
		}
	}
}
