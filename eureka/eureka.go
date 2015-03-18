// Package eureka is a service registry for the architecture configuration (nodes) as it evolves
// and passes data to edda for logging nodes and edges
package eureka

import (
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/collect"
	"github.com/adrianco/spigo/edda"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/names"
	"log"
	"sync"
	"time"
)

var Wg sync.WaitGroup

// Start eureka discovery service and set name directly
func Start(listener chan gotocol.Message, name string) {
	// use a waitgroup so whoever starts eureka can tell the logs have been flushed
	Wg.Add(1)
	defer Wg.Done()
	var msg gotocol.Message
	var ok bool
	hist := collect.NewHist(name)
	microservices := make(map[string]chan gotocol.Message, archaius.Conf.Dunbar)
	eurekaservices := make(map[string]chan gotocol.Message, 2)
	log.Println(name + ": starting")
	for {
		msg, ok = <-listener
		collect.Measure(hist, time.Since(msg.Sent))
		if !ok {
			break // channel was closed
		}
		if archaius.Conf.Msglog {
			log.Printf("%v(backlog %v): %v\n", name, len(listener), msg)
		}
		switch msg.Imposition {
		// used to wire up connections to other eureka nodes only
		case gotocol.NameDrop:
			if msg.Intention != name { // don't talk to myself
				eurekaservices[msg.Intention] = msg.ResponseChan
			}
		// for new nodes record the data, replicate and maybe pass on to be logged
		case gotocol.Put:
			if microservices[msg.Intention] == nil { // ignore duplicate requests
				microservices[msg.Intention] = msg.ResponseChan
				// replicate request
				for _, c := range eurekaservices {
					gotocol.Message{gotocol.Replicate, msg.ResponseChan, time.Now(), msg.Intention}.GoSend(c)
				}
				if edda.Logchan != nil {
					edda.Logchan <- msg
				}
			}
		case gotocol.Replicate:
			if microservices[msg.Intention] == nil { // ignore multiple requests
				microservices[msg.Intention] = msg.ResponseChan
			}
		case gotocol.Goodbye:
			close(listener)
			gotocol.Message{gotocol.Goodbye, nil, time.Now(), name}.GoSend(msg.ResponseChan)
		case gotocol.Inform:
			// don't store edges in discovery but do log them
			if edda.Logchan != nil {
				edda.Logchan <- msg
			}
		case gotocol.GetRequest:
			if msg.Intention == "" {
				log.Fatal(name + ": empty GetRequest")
			}
			if microservices[msg.Intention] != nil { // matched a unique full name
				gotocol.Message{gotocol.NameDrop, microservices[msg.Intention], time.Now(), msg.Intention}.GoSend(msg.ResponseChan)
				break
			}
			// linear scan for now, optimize if needed later by remembering timestamps or lists of services
			for n, ch := range microservices { // respond with all the names that match the service component
				// log.Printf("%v: matching %v with %v\n", name, n, msg.Intention)
				if names.Service(n) == msg.Intention {
					gotocol.Message{gotocol.NameDrop, ch, time.Now(), n}.GoSend(msg.ResponseChan)
				}
			}
		}
	}
	log.Println(name + ": closing")
}
