// package flow processes gotocol context information to collect and export request flows across the system
package flow

import (
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/gotocol"
	"os"
)

// flowmap is a map of requests of stuff, the next level of stuff is a map of spans, holding a summary
type flowmaptype map[gotocol.TraceContextType]interface{}
var flowmap flowmaptype

// file to log flow data to
var file *os.File

// Initialize the flow mapping system
func init() {
	size := archaius.Conf.Population * int(archaius.Conf.RunDuration)
	if size <= 0 {
		size = 100
	}
	flowmap = make(flowmaptype, size)
}

// Begin a new request flow
func Begin(ctx gotocol.Context, s string) {
	if file == nil {
		// do this here since Arch is not set in time for init()
		file, _ = os.Create("json_metrics/" + archaius.Conf.Arch + "_flow.json")
	}
	if flowmap[ctx.Trace] == nil {
		flowmap[ctx.Trace] = make(flowmaptype, archaius.Conf.Population)
	}
	flowmap[ctx.Trace].(flowmaptype)[ctx.Span] = ctx.String() + ":" + s
}

// Update a flow, creating it if we need to Begin a new flow
func Update(ctx gotocol.Context, s string) {
	if flowmap[ctx.Trace] == nil {
		Begin(ctx, s)
	} else {
		flowmap[ctx.Trace].(flowmaptype)[ctx.Span] = ctx.String() + ":" + s
	}
}

// Terminate a flow, flushing output and freeing the request id for re-use
func End(ctx gotocol.Context) {
	Flush(flowmap[ctx.Trace].(flowmaptype))
	delete(flowmap, ctx.Trace)
}

// Shutdown the flow mapping system and flush remaining flows
func Shutdown() {
	for _, f := range flowmap {
		Flush(f.(flowmaptype))
	}
	file.Close()
}

// Flush the spans for a request
func Flush(frequest flowmaptype) {
	a := "Trace: "
	for _, s := range frequest {
		a = a + " " + s.(string)
	}
	file.WriteString(a + "\n")
}
