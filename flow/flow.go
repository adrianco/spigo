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

// flowmap is a map of requests of stuff, the next level of stuff is a map of parents, then a map of spans, holding a map of annotations
type flowmaptype map[gotocol.TraceContextType]interface{}

// Annotation information for each step in the trace
type annotationtype struct {
	Ctx    string   `json:"ctx"`        // Context
	Host   string   `json:"host"`       // host name
	Calls  []string `json:"calls"`      // array of context:nanotimestamp
	Imp    string   `json:"imposition"` // protocol request type
	Intent string   `json:"intention"`  // request body
}

var flowmap flowmaptype

var flowlock sync.Mutex // lock changes to the maps

// file to log flow data to
var file *os.File

// Begin a new request flow
func begin(ctx gotocol.Context) {
	if !archaius.Conf.Collect {
		return
	}
	if file == nil {
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
	if flowmap[ctx.Trace] == nil {
		flowmap[ctx.Trace] = make(flowmaptype)
	}
}

// Annotate service activity on a flow and return the annotation for further use, tag = "sr" service receive, "ss" service send
func Annotate(msg gotocol.Message, tag, name string, received time.Time) *annotationtype {
	var annotation *annotationtype
	if !archaius.Conf.Collect {
		return nil
	}
	flowlock.Lock()
	if flowmap[msg.Ctx.Trace] == nil {
		begin(msg.Ctx)
	}
	if flowmap[msg.Ctx.Trace].(flowmaptype)[msg.Ctx.Parent] == nil {
		flowmap[msg.Ctx.Trace].(flowmaptype)[msg.Ctx.Parent] = make(flowmaptype)
	}
	a := (flowmap[msg.Ctx.Trace].(flowmaptype)[msg.Ctx.Parent].(flowmaptype)[msg.Ctx.Span])
	if a == nil {
		annotation = new(annotationtype)
		annotation.Host = name
		annotation.Ctx = msg.Ctx.String()
		annotation.Imp = msg.Imposition.String()
		annotation.Intent = msg.Intention
	} else {
		annotation = a.(*annotationtype)
	}
	(*annotation).Calls = append((*annotation).Calls, fmt.Sprintf("%v:%d", tag, received.UnixNano()))
	flowmap[msg.Ctx.Trace].(flowmaptype)[msg.Ctx.Parent].(flowmaptype)[msg.Ctx.Span] = annotation
	flowlock.Unlock()
	return annotation
}

// Annotate service sends on a flow, using existing annotation map and the new span
func AnnotateSend(annotation *annotationtype, span gotocol.Context) time.Time {
	now := time.Now()
	if annotation != nil {
		flowlock.Lock()
		(*annotation).Calls = append((*annotation).Calls, fmt.Sprintf("%v:%d", span.String(), now.UnixNano())) // send time
		flowlock.Unlock()
		//log.Println(*annotation)
	}
	return now
}

// Terminate a flow, flushing output and freeing the request id for re-use
func End(ctx gotocol.Context) {
	if !archaius.Conf.Collect {
		return
	}
	//	Flush(flowmap[ctx.Trace].(flowmaptype))
	//	delete(flowmap, ctx.Trace)
}

// Shutdown the flow mapping system and flush remaining flows
func Shutdown() {
	if !archaius.Conf.Collect {
		return
	}
	flowlock.Lock()
	log.Printf("Flushing flows to %v\n", file.Name())
	for _, f := range flowmap {
		Flush(f.(flowmaptype))
	}
	file.Close()
	flowlock.Unlock()
}

// Flush the spans for a request - map[parent]map[span]stuff
func Flush(trace flowmaptype) {
	//j, err := json.Marshal(trace)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//file.WriteString(fmt.Sprintf("Trace: %v\n", trace))
	Walk(trace, 0)
}

// Walk through the trace and print results in map order
func PrintWalk(flow flowmaptype) {
	for _, f := range flow {
		switch x := f.(type) {
		case string:
			fmt.Println(x)
		case flowmaptype:
			PrintWalk(x)
		case *annotationtype:
			j, _ := json.Marshal(*x)
			fmt.Printf("%v\n", string(j))
		default:
			fmt.Printf("Unknown flowmap type: %T\n", x)
		}
	}
}

// Walk through the trace and write results to file in trace order
func Walk(flow flowmaptype, parent gotocol.TraceContextType) {
	f := flow[parent] // chain through the spans in order
	switch x := f.(type) {
	case nil: // no more spans in this flow
		return
	case string:
		file.WriteString(fmt.Sprintf("%v\n", x))
	case flowmaptype:
		for s, _ := range x { // for all the spans that have this parent
			Walk(x, s)    // go in one level to print annotation
			Walk(flow, s) // chain to the next span
		}
	case *annotationtype:
		j, _ := json.Marshal(*x)
		file.WriteString(fmt.Sprintf("%v\n", string(j)))
	default:
		file.WriteString(fmt.Sprintf("Unknown flowmap type: %T\n", x))
	}
}

// common code for instrumenting requests
func Instrument(msg gotocol.Message, name string, hist *metrics.Histogram) (ann *annotationtype, span gotocol.Context) {
	received := time.Now()
	collect.Measure(hist, received.Sub(msg.Sent))
	if archaius.Conf.Msglog {
		log.Printf("%v: %v\n", name, msg)
	}
	if msg.Ctx == gotocol.NilContext {
		ann = nil
		span = gotocol.NilContext
	} else {
		ann = Annotate(msg, "sr", name, received) // annotate this request
		if ann != nil {                           // flow is enabled
			span = msg.Ctx.NewParent() // make a new context for the outbound request
		} else {
			span = gotocol.NilContext
		}
	}
	return ann, span
}
