// Package graphjson saves and loads architectures to and from graph json files
package graphjson

import (
	"encoding/json"
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// Enabled is set via command line flags to turn on json logging
var Enabled bool

var file *os.File
var edgeid int // unique id for each edge

// NodeV0r4 defines a node for version 0.4, used to make json nodes for writing
type NodeV0r4 struct {
	Node    string `json:"node"`
	Package string `json:"package"`             // name changed from 0.3 to 0.4
	Tstamp  string `json:"timestamp,omitempty"` // 0.4
}

// EdgeV0r4 defines an edge for version 0.4, used to make json edges for writing
type EdgeV0r4 struct {
	Edge   string `json:"edge"`
	Source string `json:"source"`
	Target string `json:"target"`
	Tstamp string `json:"timestamp,omitempty"` // 0.4
}

// ForgetV0r4 records an edge that has been forgotten and should be removed, forget id should match previous edge id
type ForgetV0r4 struct {
	Forget string `json:"forget"`
	Source string `json:"source"`
	Target string `json:"target"`
	Tstamp string `json:"timestamp"`
}

// DoneV0r4 records a node that goes away, and its exit status. New in 0.4
type DoneV0r4 struct {
	Done   string `json:"done"`
	Exit   string `json:"exit"`
	Tstamp string `json:"timestamp"`
}

// ElementV0r4 defines a way to read either a node, edge or done in the graph for version 0.3 or 0.4
type ElementV0r4 struct {
	Node    string `json:"node,omitempty"`
	Package string `json:"package,omitempty"`
	Service string `json:"service,omitempty"` // name changed from service 0.3 to package 0.4
	Edge    string `json:"edge,omitempty"`
	Source  string `json:"source,omitempty"`
	Target  string `json:"target,omitempty"`
	Done    string `json:"target,omitempty"`
	Exit    string `json:"exit,omitempty"`
	Tstamp  string `json:"timestamp,omitempty"`
}

// GraphV0r4 defines version 0.4 of the graphjson file format with an array of elements
type GraphV0r4 struct {
	Arch    string        `json:"arch"`
	Version string        `json:"version"`
	Args    string        `json:"args"`
	Date    string        `json:"date,omitempty"` // 0.4
	Graph   []ElementV0r4 `json:"graph"`
}

// GraphVersion extracts the version so it can be checked
type GraphVersion struct {
	Version string `json:"version"`
}

var comma bool

// Setup by opening the "arch".json file and writing a header, noting the generated architecture
// type, version and args for the run
func Setup(arch string) {
	Enabled = true
	ss := ""
	if archaius.Conf.StopStep > 0 {
		ss = fmt.Sprintf("%v", archaius.Conf.StopStep)
	}
	file, _ = os.Create("json/" + arch + ss + ".json")
	Write(fmt.Sprintf("{\n  %q:%q,\n  %q:%q,\n  %q:\"%v\",\n  %q:%q,\n  %q:[", "arch", arch, "version", "spigo-0.4", "args", os.Args, "date", time.Now().Format(time.RFC3339Nano), "graph"))
	comma = false
}

// Write a string to the file
func Write(str string) {
	file.WriteString(str)
}

// decide whether to write a comma before the newline or not
func commaNewline() string {
	if comma {
		return ",\n"
	}
	comma = true
	return "\n"
}

// WriteNode writes the node to a file given a space separated name and service type
func WriteNode(nameService string) {
	if Enabled == false {
		return
	}
	var node NodeV0r4
	fmt.Sscanf(nameService, "%s%s", &node.Node, &node.Package) // space delimited
	node.Tstamp = time.Now().Format(time.RFC3339Nano)
	// node id should be unique and service indicates service type
	nodeJSON, _ := json.Marshal(node)
	Write(fmt.Sprintf("%v    %v", commaNewline(), string(nodeJSON)))
}

// WriteEdge writes the edge to a file given a space separated from and to node name
func WriteEdge(fromTo string) {
	if Enabled == false {
		return
	}
	edgeid = edgeid + 1
	var edge EdgeV0r4
	fmt.Sscanf(fromTo, "%s%s", &edge.Source, &edge.Target) // two space delimited names
	edge.Edge = fmt.Sprintf("e%v", edgeid)
	edge.Tstamp = time.Now().Format(time.RFC3339Nano)
	edgeJSON, _ := json.Marshal(edge)
	Write(fmt.Sprintf("%v    %v", commaNewline(), string(edgeJSON)))
}

// Close completes the json file format and closes the file
func Close() {
	if Enabled == false {
		return
	}
	Write("\n  ]\n}\n")
	file.Close()
}

// ReadArch parses graphjson
func ReadArch(arch string) *GraphV0r4 {
	ss := ""
	if archaius.Conf.StopStep > 0 {
		ss = fmt.Sprintf("%v", archaius.Conf.StopStep)
	}
	fn := "json/" + arch + ss + ".json"
	log.Println("Reloading from " + fn)
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	v := new(GraphVersion)
	err = json.Unmarshal(data, v)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Version: ", v.Version)
	switch v.Version {
	case "spigo-0.3":
		fallthrough
	case "spigo-0.4":
		g := new(GraphV0r4)
		json.Unmarshal(data, g)
		log.Println("Architecture: ", g.Arch)
		return g
	default:
		log.Fatal("Uknown version ", v.Version)
		return nil
	}
}
