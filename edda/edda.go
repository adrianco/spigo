// Package edda Logs the architecture configuration (nodes and links) as it evolves
package edda

import (
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/graphml"
	"log"
	"sync"
)

// Logchan is a buffered channel for sending logging messages to, or nil if logging is off
// Created before edda starts so that messages can be buffered without depending on edda schedule
var Logchan chan gotocol.Message
var Wg sync.WaitGroup

// Start edda, to listen for logging data from services
func Start() {
	// use a waitgroup so whoever starts edda can tell the logs have been flushed
	Wg.Add(1)
	defer Wg.Done()
	var msg gotocol.Message
	var ok bool
	log.Println("edda: starting")
	if archaius.Conf.GraphmlFile != "" {
		graphml.Enabled = true
	}
	if archaius.Conf.GraphjsonFile != "" {
		graphjson.Enabled = true
	}
	graphml.Setup(archaius.Conf.GraphmlFile)
	graphjson.Setup(archaius.Conf.GraphjsonFile)
	for {
		msg, ok = <-Logchan
		if !ok {
			break // channel was closed
		}
		if archaius.Conf.Msglog {
			log.Printf("edda(backlog %v): %v\n", len(Logchan), msg)
		}
		if msg.Imposition == gotocol.Inform {
			graphml.WriteEdge(msg.Intention)
			graphjson.WriteEdge(msg.Intention)
		} else {
			if msg.Imposition == gotocol.Hello {
				graphml.WriteNode(msg.Intention)
				graphjson.WriteNode(msg.Intention)
			}
		}
	}
	log.Println("edda: closing")
	graphml.Close()
	graphjson.Close()
}
