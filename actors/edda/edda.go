// Package edda Logs the architecture configuration (nodes and links) as it evolves
package edda

import (
	"github.com/adrianco/spigo/tooling/archaius"
	"github.com/adrianco/spigo/tooling/collect"
	"github.com/adrianco/spigo/tooling/gotocol"
	"github.com/adrianco/spigo/tooling/graphjson"
	"github.com/adrianco/spigo/tooling/graphml"
	"github.com/adrianco/spigo/tooling/graphneo4j"
	"github.com/adrianco/spigo/tooling/names"
	"log"
	"strings"
	"sync"
	"time"
)

// Logchan is a buffered channel for sending logging messages to, or nil if logging is off
// Created before edda starts so that messages can be buffered without depending on edda schedule
var Logchan chan gotocol.Message

// Wg is a waitgroup so whoever starts edda can wait for it to finish logging before it shuts down
var Wg sync.WaitGroup

// Start edda, to listen for logging data from services
func Start(name string) {
	// use a waitgroup so whoever starts edda can tell the logs have been flushed
	Wg.Add(1)
	defer Wg.Done()
	if Logchan == nil {
		return
	}
	var msg gotocol.Message
	microservices := make(map[string]bool, archaius.Conf.Dunbar)
	edges := make(map[string]bool, archaius.Conf.Dunbar)
	var ok bool
	hist := collect.NewHist(name)
	log.Println(name + ": starting")
	if archaius.Conf.GraphmlFile != "" {
		graphml.Setup(archaius.Conf.GraphmlFile)
	}
	if archaius.Conf.GraphjsonFile != "" {
		graphjson.Setup(archaius.Conf.GraphjsonFile)
	}
	if archaius.Conf.Neo4jURL != "" {
		graphneo4j.Setup(archaius.Conf.Neo4jURL)
	}
	for {
		msg, ok = <-Logchan
		collect.Measure(hist, time.Since(msg.Sent))
		if !ok {
			break // channel was closed
		}
		if archaius.Conf.Msglog {
			log.Printf("%v(backlog %v): %v\n", name, len(Logchan), msg)
		}
		switch msg.Imposition {
		case gotocol.Inform:
			edge := names.FilterEdge(msg.Intention)
			if edges[edge] == false { // only log an edge once
				edges[edge] = true
				graphml.WriteEdge(edge)
				graphjson.WriteEdge(edge, msg.Sent)
				graphneo4j.WriteEdge(strings.Replace(msg.Intention, "-", "_", -1), msg.Sent)
			}
		case gotocol.Put:
			node := names.FilterNode(msg.Intention)
			if microservices[node] == false { // only log a node once
				microservices[node] = true
				graphml.WriteNode(node + " " + names.Package(msg.Intention))
				graphjson.WriteNode(node+" "+names.Package(msg.Intention), msg.Sent)
				graphneo4j.WriteNode(strings.Replace(msg.Intention, "-", "_", -1)+" "+names.Package(msg.Intention), msg.Sent)
			}
		case gotocol.Forget: // forget the edge
			// problem here in that edges may be reported multiple times from several sources
			// however after filtering all matching edges are reported as forgotten when the first is
			// need to maintain the full model and a filtered model with counts
			edge := names.FilterEdge(msg.Intention)
			if edges[edge] == true { // only remove an edge once
				edges[edge] = false
				graphjson.WriteForget(edge, msg.Sent)
			}
		case gotocol.Delete: // remove the node
			node := names.FilterNode(msg.Intention)
			if microservices[node] == true { // only remove nodes that exist, and only log it once
				microservices[node] = false
				graphjson.WriteDone(node, msg.Sent)
			}
		}
	}
	log.Println(name + ": closing")
	graphml.Close()
	graphjson.Close()
	graphneo4j.Close()
}
