// functions for writing graphml

package graphml

import (
	"os"
	"fmt"
)

var Enabled bool
var file *os.File
var edgeid int	// unique id for each edge to keep graphml happy

// write the header to the file
func Setup() {
	if Enabled == false {
		return
	}
        file, _ = os.Create("spigo.graphml")
        file.WriteString(
"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n  <graphml xmlns=\"http://graphml.graphdrawing.org/xmlns/graphml\"\n   xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"\n   xsi:schemaLocation=\"http://graphml.graphdrawing.org/xmlns/graphml http://www.yworks.com/xml/schema/graphml/1.0/ygraphml.xsd\"\n    xmlns:y=\"http://www.yworks.com/xml/graphml\">\n    <key id=\"d0\" for=\"node\" yfiles.type=\"nodegraphics\"/>\n    <key id=\"d1\" for=\"edge\" yfiles.type=\"edgegraphics\"/>\n    <graph id=\"spigo\" edgedefault=\"directed\">\n")
}

func Node(name string) {
        if Enabled == false {
                return
        }
	file.WriteString(fmt.Sprintf("      <node id=\"%v\"/>\n", name))
}

func Edge(from, to string) {
        if Enabled == false {
                return
        }
	file.WriteString(fmt.Sprintf("      <edge id=\"e%v\" source=\"%v\" target=\"%v\"/>\n", edgeid, from, to))
	edgeid = edgeid + 1
}

func Close() {
        if Enabled == false {
                return
        }
	file.WriteString("    </graph>\n  </graphml>\n")
	file.Close()
}

