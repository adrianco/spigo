// Logs the architecture (nodes and links) as it evolves
package logger

import (
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/graphml"
	"log"
)

var Logchan chan gotocol.Message
var Msglog bool

// separate goroutine to gather logging data from pirates
func GoLog(arch string) {
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
