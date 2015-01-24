// Package graphml writes nodes and edges to spigo.graphml in the XML based GraphML format used by
// the freely available yEdit tool from yfiles.com
package graphml

import (
	"fmt"
	"os"
)

// Enabled is set by command line flags to turn on graphml logging
var Enabled bool

var file *os.File
var edgeid int // unique id for each edge to keep graphml happy

// Setup opens the file and and writes the header
func Setup() {
	if Enabled == false {
		return
	}
	file, _ = os.Create("spigo.graphml")
	file.WriteString(
		"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n  <graphml xmlns=\"http://graphml.graphdrawing.org/xmlns/graphml\"\n   xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"\n   xsi:schemaLocation=\"http://graphml.graphdrawing.org/xmlns/graphml http://www.yworks.com/xml/schema/graphml/1.0/ygraphml.xsd\"\n    xmlns:y=\"http://www.yworks.com/xml/graphml\">\n    <key id=\"d0\" for=\"node\" yfiles.type=\"nodegraphics\"/>\n    <key id=\"d1\" for=\"edge\" yfiles.type=\"edgegraphics\"/>\n    <key id=\"d2\" for=\"node\" attr.name=\"Text\" attr.type=\"string\"/>\n    <graph id=\"spigo\" edgedefault=\"directed\">\n")
}

// WriteNode logs a node in the file given a space separated name and service type
func WriteNode(nameService string) {
	if Enabled == false {
		return
	}
	var name, service string
	fmt.Sscanf(nameService, "%s%s", &name, &service) // space delimited
	// node name should be unique and service indicates service type
	file.WriteString(fmt.Sprintf("      <node id=\"%v\"><data key=\"service\">%v</data></node>\n", name, service))
}

func edge(from, to string) string {
	if Enabled == false {
		return ""
	}
	edgeid = edgeid + 1
	return fmt.Sprintf("      <edge id=\"e%v\" source=\"%v\" target=\"%v\"/>\n", edgeid, from, to)
}

// Write a string to the file
func Write(str string) {
	file.WriteString(str)
}

// WriteEdge logs and edge in the file given a space separated from and to name
func WriteEdge(fromTo string) {
	if Enabled == false {
		return
	}
	var from, to string
	fmt.Sscanf(fromTo, "%s%s", &from, &to) // two space delimited names
	file.WriteString(edge(from, to))
}

// Close finishes off the file footer and closes it
func Close() {
	if Enabled == false {
		return
	}
	file.WriteString("    </graph>\n  </graphml>\n")
	file.Close()
}
