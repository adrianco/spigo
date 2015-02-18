// Package eureka is a service registry for the architecture configuration (nodes) as it evolves
// and passes data to edda for logging nodes and edges
package eureka

import (
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/collect"
	"github.com/adrianco/spigo/edda"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
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
	servicetypes := make(map[string]graphjson.NodeV0r3, archaius.Conf.Dunbar) // service type for each microservice
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
		// for new nodes and edges record the data and maybe pass on to be logged
		case gotocol.Hello:
			var node graphjson.NodeV0r3
			fmt.Sscanf(msg.Intention, "%s%s", &node.Node, &node.Service) // space delimited
			microservices[node.Node] = msg.ResponseChan
			servicetypes[node.Node] = node
			if edda.Logchan != nil {
				edda.Logchan <- msg
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
			if msg.Intention != "" && microservices[msg.Intention] != nil {
				gotocol.Message{gotocol.GetResponse, listener, time.Now(), servicetypes[msg.Intention].Service}.GoSend(msg.ResponseChan)
			}
		}
	}
	log.Println(name + ": closing")
}
