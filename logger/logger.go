// Logs the architecture (nodes and links) as it evolves

package logger

import (
	"log"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/graphml"
)

var Logchan chan gotocol.Message

// separate goroutine to gather logging data from pirates
func GoLog(arch string) {
	var msg gotocol.Message
	var ok bool
	Logchan = make(chan gotocol.Message, 100) // buffered channel
        log.Println("Logger starting")
	graphml.Setup()
	graphjson.Setup(arch)
	for {
		msg, ok = <-Logchan
		//log.Printf("logger:%v %v\n", len(Logchan), msg)
		if !ok {
			break // channel was closed
		}
		if msg.Imposition == gotocol.Inform {
			graphml.WriteEdge(msg.Intention)
			graphjson.WriteEdge(msg.Intention)
		}
	}
	log.Println("Logger closing")
	graphml.Close()
	graphjson.Close()
}
