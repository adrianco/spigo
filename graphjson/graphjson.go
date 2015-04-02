// Package graphjson saves and loads architectures to and from graph json files
package graphjson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// Enabled is set via command line flags to turn on json logging
var Enabled bool

var file *os.File
var edgeid int // unique id for each edge

// NodeV0r3 defines a node for version 0.3, used to make json nodes for writing
type NodeV0r3 struct {
	Node    string `json:"node"`
	Service string `json:"service"`
}

// EdgeV0r3 defines an edge for version 0.3, used to make json edges for writing
type EdgeV0r3 struct {
	Edge   string `json:"edge"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// ElementV0r3 defines a way to read either a node or an edge in the graph for version 0.3
type ElementV0r3 struct {
	Node    string `json:"node,omitempty"`
	Service string `json:"service,omitempty"`
	Edge    string `json:"edge,omitempty"`
	Source  string `json:"source,omitempty"`
	Target  string `json:"target,omitempty"`
}

// GraphV0r3 defines version 0.3 of the graphjson file format with an array of elements
type GraphV0r3 struct {
	Arch    string        `json:"arch"`
	Version string        `json:"version"`
	Args    string        `json:"args"`
	Graph   []ElementV0r3 `json:"graph"`
}

// GraphVersion extracts the version so it can be checked
type GraphVersion struct {
	Version string `json:"version"`
}

var comma bool

// Setup by opening the "arch".json file and writing a header, noting the generated architecture
// type, version and args for the run
func Setup(arch string) {
	if Enabled == false {
		return
	}
	file, _ = os.Create("json/" + arch + ".json")
	Write(fmt.Sprintf("{\n  \"arch\":\"%v\",\n  \"version\":\"spigo-0.3\",\n  \"args\":\"%v\",\n  \"graph\":[", arch, os.Args))
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
	var node NodeV0r3
	fmt.Sscanf(nameService, "%s%s", &node.Node, &node.Service) // space delimited
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
	var edge EdgeV0r3
	fmt.Sscanf(fromTo, "%s%s", &edge.Source, &edge.Target) // two space delimited names
	edge.Edge = fmt.Sprintf("e%v", edgeid)
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
func ReadArch(arch string) *GraphV0r3 {
	data, err := ioutil.ReadFile("json/" + arch + ".json")
	if err != nil {
		log.Fatal(err)
	}
	v := new(GraphVersion)
	json.Unmarshal(data, v)
	log.Println("Version: ", v.Version)
	switch v.Version {
	case "spigo-0.3":
		g := new(GraphV0r3)
		json.Unmarshal(data, g)
		log.Println("Architecture: ", g.Arch)
		return g
	default:
		log.Println("Uknown version ", v.Version)
		return nil
	}
}
