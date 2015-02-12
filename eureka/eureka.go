// Package eureka is a service registry for the architecture configuration (nodes) as it evolves
// and passes data to edda for logging nodes and edges
package eureka

import (
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/edda"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"log"
)

// Start eureka discovery service for an architecture with an initial scale
func Start(listener chan gotocol.Message) {
	var msg gotocol.Message
	var ok bool
	microservices := make(map[string]chan gotocol.Message, archaius.Conf.Dunbar)
	servicetypes := make(map[string]graphjson.NodeV0r3, archaius.Conf.Dunbar) // service type for each microservice
	log.Println("eureka: starting")
	for {
		msg, ok = <-listener
		if !ok {
			break // channel was closed
		}
		if archaius.Conf.Msglog {
			log.Printf("eureka(backlog %v): %v\n", len(listener), msg)
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
			if edda.Logchan != nil {
				close(edda.Logchan)
			}
			gotocol.Message{gotocol.Goodbye, nil, ""}.GoSend(msg.ResponseChan)
			break
		case gotocol.Inform:
			// don't store edges in discovery but do log them
			if edda.Logchan != nil {
				edda.Logchan <- msg
			}
		case gotocol.GetRequest:
			if msg.Intention != "" && microservices[msg.Intention] != nil {
				gotocol.Message{gotocol.GetResponse, listener, servicetypes[msg.Intention].Service}.GoSend(msg.ResponseChan)
			}
		}
	}
	log.Println("eureka: closing")
}
