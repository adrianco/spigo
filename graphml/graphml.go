// The graphml package provides functions for writing GraphML.
package graphml

import (
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
)

var Enabled bool

var (
	// fileMutex guards writes to the file - concurrent writes
	// to a file do not always atomically append on all platforms.
	fileMutex sync.Mutex
	file      *os.File
)

var edgeid uint32 // unique id for each edge to keep graphml happy

const header = `<?xml version="1.0" encoding="UTF-8"?>
  <graphml xmlns="http://graphml.graphdrawing.org/xmlns/graphml"
   xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
   xsi:schemaLocation="http://graphml.graphdrawing.org/xmlns/graphml http://www.yworks.com/xml/schema/graphml/1.0/ygraphml.xsd"
    xmlns:y="http://www.yworks.com/xml/graphml">
    <key id="d0" for="node" yfiles.type="nodegraphics"/>
    <key id="d1" for="edge" yfiles.type="edgegraphics"/>
    <graph id="spigo" edgedefault="directed">`

// Setup creates the output file and writes the header to it.
func Setup() {
	if !Enabled {
		return
	}
	var err error
	file, err = os.Create("spigo.graphml")
	if err != nil {
		log.Fatal(err)
	}
	printf("%s", header)
}

func Node(name string) {
	if !Enabled {
		return
	}
	printf(`      <node id="%v"/>`, name)
}

func Edge(from, to string) {
	if !Enabled {
		return
	}
	id := atomic.AddUint32(&edgeid, 1)
	printf(`      <edge id="e%v" source="%v" target="%v"/>`, id, from, to)
}

func Close() {
	if file == nil {
		return
	}
	printf("    </graph>\n  </graphml>")
	file.Close()
	file = nil
}

func printf(f string, a ...interface{}) {
	fileMutex.Lock()
	fmt.Fprintln(file, fmt.Sprintf(f, a...))
	fileMutex.Unlock()
}
