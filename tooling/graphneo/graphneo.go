// Package graphneo saves architectures to neo4j
package graphneo

import (
	"database/sql"
	"fmt"
	"github.com/adrianco/spigo/tooling/archaius"
	"github.com/adrianco/spigo/tooling/dhcp"
	"github.com/adrianco/spigo/tooling/names"
	_ "gopkg.in/cq.v1"
	"log"
	"os"
	"time"
)

// Enabled is set via command line flags to turn on neo logging
var Enabled bool
var db *sql.DB
var ss string

// Setup by opening the "arch".json file and writing a header, noting the generated architecture
// type, version and args for the run
func Setup(arch string) {
	Enabled = true
	if archaius.Conf.StopStep > 0 {
		ss = fmt.Sprintf("%v", archaius.Conf.StopStep)
	}
	tmp, err := sql.Open("neo4j-cypher", "http://neo4j:"+os.Getenv("NEO4JPASSWORD")+"@localhost:7474")
	if err != nil {
		log.Fatal(err)
	}
	db = tmp
}

// Write an entry to the database
func Write(str string) {
	//log.Println(str)
	stmt, err := db.Prepare(str)
	if err != nil {
		log.Fatal(err)
	}
	stmt.Exec("")
	log.Println(stmt)
	stmt.Close()
}

// WriteNode writes the node to a file given a space separated name and service type
func WriteNode(nameService string, t time.Time) {
	if Enabled == false {
		return
	}
	var node, pack string
	fmt.Sscanf(nameService, "%s%s", &node, &pack) // space delimited
	tstamp := t.Format(time.RFC3339Nano)
	// node id should be unique and package indicates service type
	Write(fmt.Sprintf("CREATE (%v_%v:Node {name:%q, package:%q, timestamp:%q, ip:%q, region:%q, zone:%q})\nRETURN %v_%v", archaius.Conf.Arch+ss, names.Instance(node), node, pack, tstamp, dhcp.Lookup(node), names.Region(node), names.Zone(node), archaius.Conf.Arch+ss, names.Instance(node)))
}

// WriteDone records that a node has gone away normally
//func WriteDone(name string, t time.Time) {
//	if Enabled == false {
//		return
//	}
//}

// WriteEdge writes the edge to a file given a space separated from and to node name
func WriteEdge(fromTo string, t time.Time) {
	if Enabled == false {
		return
	}
	var source, target string
	fmt.Sscanf(fromTo, "%s%s", &source, &target) // two space delimited names
	tstamp := t.Format(time.RFC3339Nano)
	Write(fmt.Sprintf("CREATE (%v_%v)-[:CONNECTION {timestamp:%q}]->(%v_%v)\n", archaius.Conf.Arch+ss, names.Instance(source), tstamp, archaius.Conf.Arch+ss, names.Instance(target)))
}

// WriteForget writes the forgotten edge to a file given a space separated edge id, from and to node names
//func WriteForget(fromTo string, t time.Time) {
//	if Enabled == false {
//		return
//	}
//}

// Close the database session
func Close() {
	if Enabled == false {
		return
	}
	db.Close()
}
