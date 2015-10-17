// package flow processes gotocol context information to collect and export request flows across the system
package flow

import (
	"encoding/json"
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/collect"
	"github.com/adrianco/spigo/gotocol"
	"github.com/codahale/metrics"
	"log"
	"os"
	"sync"
	"time"
)

// value for zipkin span direction
type Values int

const (
	CS      Values = iota // client send
	SR                    // server receive
	SS                    // server send
	CR                    // client receive
	Unknown               // something went wrong
)

// pretty printer for Values
func (v Values) String() string {
	switch v {
	case CS:
		return "cs"
	case SR:
		return "sr"
	case SS:
		return "ss"
	case CR:
		return "cr"
	default:
		return "unknown"
	}
}

// flowmap is a map by traceid of slices of pointers to spannotations
type flowmaptype map[gotocol.TraceContextType][]*spannotype

// Annotation information for each step in the span
type spannotype struct {
	Ctx       string `json:"ctx"`        // Context
	Host      string `json:"host"`       // host name
	Imp       string `json:"imposition"` // protocol request type
	Intent    string `json:"intention"`  // request body
	Timestamp int64  `json:"ts"`         // unix nanotimestamp
	Value     string `json:"value"`      // direction of span
}

var flowmap flowmaptype

var flowlock sync.Mutex // lock changes to the maps

// file to log flow data to
var file *os.File

// setup and initialize the flow log
func setup() {
	// do this here since Arch is not set in time for init()
	f, err := os.Create("json_metrics/" + archaius.Conf.Arch + "_flow.json")
	if err != nil {
		log.Fatal(err)
	} else {
		file = f
	}
	// Initialize the flow mapping system
	flowmap = make(flowmaptype, archaius.Conf.Population)
}

// Common Annotation code
func annotate(msg gotocol.Message, name string, t time.Time, resp, others Values) *spannotype {
	if file == nil {
		setup()
	}
	if flowmap[msg.Ctx.Trace] == nil {
		flowmap[msg.Ctx.Trace] = make([]*spannotype, 0, 2) // reserve space for at least 2 annotations in a span
	}
	annotation := new(spannotype)
	annotation.Host = name
	annotation.Ctx = msg.Ctx.String()
	annotation.Imp = msg.Imposition.String()
	annotation.Intent = msg.Intention
	annotation.Timestamp = t.UnixNano()
	if msg.Imposition == gotocol.GetResponse {
		annotation.Value = resp.String()
	} else {
		annotation.Value = others.String()
	}
	return annotation
}

// Annotate service activity when receiving a message
func AnnotateReceive(msg gotocol.Message, name string, received time.Time) {
	if !archaius.Conf.Collect {
		return
	}
	flowlock.Lock()
	flowmap[msg.Ctx.Trace] = append(flowmap[msg.Ctx.Trace], annotate(msg, name, received, CR, SR))
	flowlock.Unlock()
	return
}

// Annotate service sends on a flow
func AnnotateSend(msg gotocol.Message, name string) {
	if !archaius.Conf.Collect {
		return
	}
	flowlock.Lock()
	flowmap[msg.Ctx.Trace] = append(flowmap[msg.Ctx.Trace], annotate(msg, name, msg.Sent, SS, CS))
	flowlock.Unlock()
	return
}

// Terminate a flow, flushing output and freeing the request id to keep the map smaller
func End(ctx gotocol.Context) {
	if !archaius.Conf.Collect {
		return
	}
	Flush(flowmap[ctx.Trace])
	delete(flowmap, ctx.Trace)
}

// Shutdown the flow mapping system and flush remaining flows
func Shutdown() {
	if !archaius.Conf.Collect {
		return
	}
	flowlock.Lock()
	log.Printf("Flushing flows to %v\n", file.Name())
	for _, f := range flowmap {
		Flush(f)
	}
	file.Close()
	flowlock.Unlock()
}

// Flush the spans for a request
func Flush(trace []*spannotype) {
	for _, a := range trace {
		j, err := json.Marshal(*a)
		if err != nil {
			log.Fatal(err)
		}
		file.WriteString(fmt.Sprintf("%v\n", string(j)))
		//log.Println(string(j))
	}
}

// common code for instrumenting requests
func Instrument(msg gotocol.Message, name string, hist *metrics.Histogram) {
	received := time.Now()
	collect.Measure(hist, received.Sub(msg.Sent))
	if archaius.Conf.Msglog {
		log.Printf("%v: %v\n", name, msg)
	}
	if msg.Ctx != gotocol.NilContext {
		AnnotateReceive(msg, name, received) // store the annotation for this request
	}
}
