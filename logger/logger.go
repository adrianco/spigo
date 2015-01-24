// Package logger Logs the architecture (nodes and links) as it evolves
package logger

import (
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/graphml"
	"log"
)

// Logchan is a buffered channel for sending logging messages to, or nil if logging is off
var Logchan chan gotocol.Message

// Msglog if true, log each message received on the console
var Msglog bool

// Start the logger, to listen for logging data from pirates
func Start(arch string) {
	var msg gotocol.Message
	var ok bool
	Logchan = make(chan gotocol.Message, 100) // buffered channel
	log.Println("logger: starting")
	graphml.Setup()
	graphjson.Setup(arch)
	for {
		msg, ok = <-Logchan
		if !ok {
			break // channel was closed
		}
		if Msglog {
			log.Printf("logger(backlog %v): %v\n", len(Logchan), msg)
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
	log.Println("logger: closing")
	graphml.Close()
	graphjson.Close()
}
