// functions for writing graphml

package graphjson

import (
	"fmt"
	"os"
)

var Enabled bool
var file *os.File
var edgeid int // unique id for each edge
var comma bool

// write the header to the file
func Setup() {
	if Enabled == false {
		return
	}
	file, _ = os.Create("spigo.json")
	file.WriteString("{\n  \"version\":\"spigo-0.0\",\n  \"graph\":[")
	comma = false
}

func Write(str string) {
	file.WriteString(str)
}

func commaNewline() string {
	if comma {
		return ",\n"
	} else {
		comma = true
		return "\n"
	}
}

func WriteNode(name string) {
	if Enabled == false {
		return
	}
	// node id should be unique but name is an arbitrary label set the same for now
	Write(fmt.Sprintf("%v    { \"node\":\"%v\", \"name\":\"%v\" }", commaNewline(), name, name))
}

func Edge(from, to string) string {
	if Enabled == false {
		return ""
	}
	edgeid = edgeid + 1
	return fmt.Sprintf("%v    { \"edge\":\"e%v\", \"source\":\"%v\", \"target\":\"%v\" }", commaNewline(), edgeid, from, to)
}

func WriteEdge(from, to string) {
	if Enabled == false {
		return
	}
	Write(Edge(from, to))
}

func Close() {
	if Enabled == false {
		return
	}
	Write("\n  ]\n}\n")
	file.Close()
}
