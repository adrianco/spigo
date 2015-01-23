// functions for writing graphml

package graphml

import (
	"fmt"
	"os"
)

var Enabled bool
var file *os.File
var edgeid int // unique id for each edge to keep graphml happy

// write the header to the file
func Setup() {
	if Enabled == false {
		return
	}
	file, _ = os.Create("spigo.graphml")
	file.WriteString(
		"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n  <graphml xmlns=\"http://graphml.graphdrawing.org/xmlns/graphml\"\n   xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"\n   xsi:schemaLocation=\"http://graphml.graphdrawing.org/xmlns/graphml http://www.yworks.com/xml/schema/graphml/1.0/ygraphml.xsd\"\n    xmlns:y=\"http://www.yworks.com/xml/graphml\">\n    <key id=\"d0\" for=\"node\" yfiles.type=\"nodegraphics\"/>\n    <key id=\"d1\" for=\"edge\" yfiles.type=\"edgegraphics\"/>\n    <key id=\"d2\" for=\"node\" attr.name=\"Text\" attr.type=\"string\"/>\n    <graph id=\"spigo\" edgedefault=\"directed\">\n")
}

func WriteNode(serviceName string) {
	if Enabled == false {
		return
	}
	var name, service string
	fmt.Sscanf(serviceName, "%s%s", &name, &service) // space delimited
	// node name should be unique and service indicates service type
	file.WriteString(fmt.Sprintf("      <node id=\"%v\"><data key=\"service\">%v</data></node>\n", name, service))
}

func Edge(from, to string) string {
	if Enabled == false {
		return ""
	}
	edgeid = edgeid + 1
	return fmt.Sprintf("      <edge id=\"e%v\" source=\"%v\" target=\"%v\"/>\n", edgeid, from, to)
}

func Write(str string) {
	file.WriteString(str)
}

func WriteEdge(fromTo string) {
	if Enabled == false {
		return
	}
	var from, to string
	fmt.Sscanf(fromTo, "%s%s", &from, &to) // two space delimited names
	file.WriteString(Edge(from, to))
}

func Close() {
	if Enabled == false {
		return
	}
	file.WriteString("    </graph>\n  </graphml>\n")
	file.Close()
}
