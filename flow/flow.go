// package flow processes gotocol context information to collect and export request flows across the system
package flow

import (
	//"encoding/json"
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/gotocol"
	"log"
	"os"
)

// flowmap is a map of requests of stuff, the next level of stuff is a map of spans, holding a summary
type flowmaptype map[gotocol.TraceContextType]interface{}

var flowmap flowmaptype

// file to log flow data to
var file *os.File

// Begin a new request flow
func Begin(ctx gotocol.Context) {
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

// Update a flow, creating it if we need to Begin a new flow
func Update(ctx gotocol.Context, s string) {
	if !archaius.Conf.Collect {
		return
	}
	if flowmap[ctx.Trace] == nil {
		Begin(ctx)
	}
	if flowmap[ctx.Trace].(flowmaptype)[ctx.Parent] == nil {
		flowmap[ctx.Trace].(flowmaptype)[ctx.Parent] = make(flowmaptype)
	}
	flowmap[ctx.Trace].(flowmaptype)[ctx.Parent].(flowmaptype)[ctx.Span] = ctx.String() + "-" + s
}

// Terminate a flow, flushing output and freeing the request id for re-use
func End(ctx gotocol.Context) {
	if !archaius.Conf.Collect {
		return
	}
	Flush(flowmap[ctx.Trace].(flowmaptype))
	delete(flowmap, ctx.Trace)
}

// Shutdown the flow mapping system and flush remaining flows
func Shutdown() {
	if !archaius.Conf.Collect {
		return
	}
	for _, f := range flowmap {
		Flush(f.(flowmaptype))
	}
	file.Close()
}

// Flush the spans for a request - map[parent]map[span]stuff
func Flush(trace flowmaptype) {
	//j, err := json.Marshal(trace)
	//if err != nil {
	//	log.Fatal(err)
	//}
	file.WriteString(fmt.Sprintf("Trace: %v\n", trace))
}
