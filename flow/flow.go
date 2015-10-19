// package flow processes gotocol context information to collect and export request flows across the system
package flow

import (
	"encoding/json"
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/collect"
	"github.com/adrianco/spigo/dhcp"
	"github.com/adrianco/spigo/gotocol"
	"github.com/codahale/metrics"
	"log"
	"os"
	"sort"
	"strings"
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
	Ctx       string `json:"ctx"`        // Context as string
	Host      string `json:"host"`       // host name
	Imp       string `json:"imposition"` // protocol request type
	Intent    string `json:"intention"`  // request body
	Timestamp int64  `json:"ts"`         // unix nanotimestamp
	Value     string `json:"value"`      // direction of span
}

// sortable spans
type ByCtx []*spannotype

func (a ByCtx) Len() int             { return len(a) }
func (a ByCtx) Swap(i, j int)        { a[i], a[j] = a[j], a[i] }
func (a ByCtx) Less(i, j int) bool { // sort by span first then time
	if a[i].Ctx == a[j].Ctx {
		return a[i].Timestamp < a[j].Timestamp
	} else {
		return a[i].Ctx < a[j].Ctx
	}
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
	file.WriteString("[\n")
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
	//Flush(ctx.Trace, flowmap[ctx.Trace])
	//delete(flowmap, ctx.Trace)
}

// Shutdown the flow mapping system and flush remaining flows
func Shutdown() {
	if !archaius.Conf.Collect {
		return
	}
	flowlock.Lock()
	log.Printf("Flushing flows to %v\n", file.Name())
	comma := false
	for c, f := range flowmap {
		if comma {
			file.WriteString(",\n")
		} else {
			comma = true
		}
		Flush(c, f)
	}
	file.WriteString("]\n")
	file.Close()
	flowlock.Unlock()
}

/* example: Zipkin format is an array of these
{
  "traceId": "5e27c67030932221",
  "name": "GET",
  "id": "38357d8f309b379d",
  "parentId": "5e27c67030932221",
  "annotations": [
    {
      "endpoint": {
        "serviceName": "zipkin-query",
        "ipv4": "127.0.0.1"
      },
      "timestamp": 1444780030334000,
      "value": "cs"
    },
    {
      "endpoint": {
        "serviceName": "zipkin-query",
        "ipv4": "172.17.0.84"
      },
      "timestamp": 1444780030643000,
      "value": "sr"
    },
    {
      "endpoint": {
        "serviceName": "zipkin-query",
        "ipv4": "127.0.0.1"
      },
      "timestamp": 1444780031521000,
      "value": "cr"
    },
    {
      "endpoint": {
        "serviceName": "zipkin-query",
        "ipv4": "172.17.0.84"
      },
      "timestamp": 1444780031689000,
      "value": "ss"
    }
  ]
},
*/

// endpoint for zipkin
type zipkinendpoint struct {
	Servicename string `json:"servicename"`
	Ipv4        string `json:"ipv4"`
	Port        int    `json:"port"`
}

// annotation for zipkin
type zipkinannotation struct {
	Endpoint  zipkinendpoint `json:"endpoint"`
	Timestamp int64          `json:"timestamp"`
	Value     string         `json:"value"`
}

// trace for zipkin
type zipkinspan struct {
	Traceid     string             `json:"traceId"`
	Name        string             `json:"name"`
	Id          string             `json:"id"`
	ParentId    string             `json:"parentId"`
	Debug       string             `json:"debug"`
	Annotations []zipkinannotation `json:"annotations"`
}

func WriteZip(zip zipkinspan) {
	j, err := json.Marshal(zip)
	if err != nil {
		log.Fatal(err)
	}
	file.WriteString(fmt.Sprintf("%v\n", string(j)))
}

// Flush the spans for a request in zipkin format
func Flush(t gotocol.TraceContextType, trace []*spannotype) {
	var zip zipkinspan
	var ctx string
	n := -1
	sort.Sort(ByCtx(trace))
	for _, a := range trace {
		//fmt.Println(*a)
		if ctx != a.Ctx { // new span
			if ctx != "" { // not the first
				WriteZip(zip)
				file.WriteString(",")
				zip.Annotations = nil
			}
			n++
			zip.Traceid = fmt.Sprintf("%v", t)
			zip.Name = a.Imp
			s := strings.SplitAfter(a.Ctx, "s")                            // tXpYsZ -> [tXpYs, Z]
			p := strings.TrimSuffix(strings.SplitAfter(s[0], "p")[1], "s") // tXpYs -> [tXp, Ys] -> Ys -> Y
			zip.Id = s[1]
			zip.ParentId = p
			zip.Debug = "false"
			ctx = a.Ctx
		}
		var ann zipkinannotation
		ann.Endpoint.Servicename = a.Host
		ann.Endpoint.Ipv4 = dhcp.Lookup(a.Host)
		ann.Endpoint.Port = 8080
		ann.Timestamp = a.Timestamp / 1000 // convert from UnixNano to Microseconds
		ann.Value = a.Value
		zip.Annotations = append(zip.Annotations, ann)
	}
	WriteZip(zip)
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
