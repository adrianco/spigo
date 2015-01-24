// Package graphjson writes a json representation of the spigo network of nodes and edges
// to spigo.json
package graphjson

// Imports
import (
	"fmt"
	"os"
)

// Enabled is set via command line flags to turn on json logging
var Enabled bool

var file *os.File
var edgeid int // unique id for each edge
var comma bool

// Setup by opening the spigo.json file and writing a header, noting the generated architecture
// type, version and args for the run
func Setup(arch string) {
	if Enabled == false {
		return
	}
	file, _ = os.Create("spigo.json")
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
	} else {
		comma = true
		return "\n"
	}
}

// WriteNote writes the node to a file given a space separated name and service type
func WriteNode(nameService string) {
	if Enabled == false {
		return
	}
	var name, service string
	fmt.Sscanf(nameService, "%s%s", &name, &service) // space delimited
	// node id should be unique and service indicates service type
	Write(fmt.Sprintf("%v    { \"node\":\"%v\", \"service\":\"%v\" }", commaNewline(), name, service))
}

func edge(from, to string) string {
	if Enabled == false {
		return ""
	}
	edgeid = edgeid + 1
	return fmt.Sprintf("%v    { \"edge\":\"e%v\", \"source\":\"%v\", \"target\":\"%v\" }", commaNewline(), edgeid, from, to)
}

// WriteEdge writes the edge to a file given a space separated from and to node name
func WriteEdge(fromTo string) {
	if Enabled == false {
		return
	}
	var from, to string
	fmt.Sscanf(fromTo, "%s%s", &from, &to) // two space delimited names
	Write(edge(from, to))
}

// Close completes the json file format and closes the file
func Close() {
	if Enabled == false {
		return
	}
	Write("\n  ]\n}\n")
	file.Close()
}
