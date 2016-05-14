// Package graphneo4j saves architectures to neo4j
package graphneo4j

import (
	"database/sql"
	"fmt"
	"github.com/adrianco/spigo/tooling/archaius"
	"github.com/adrianco/spigo/tooling/dhcp"
	"github.com/adrianco/spigo/tooling/gotocol"
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
var epoch int64

// Setup by opening the "arch".json file and writing a header, noting the generated architecture
// type, version and args for the run
func Setup(neo4jurl string) {
	Enabled = true
	if archaius.Conf.StopStep > 0 {
		ss = fmt.Sprintf("%v", archaius.Conf.StopStep)
	}
	tmp, err := sql.Open("neo4j-cypher", "http://neo4j:"+os.Getenv("NEO4JPASSWORD")+"@"+neo4jurl)
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
	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}
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
	nodestmt, err := db.Prepare(fmt.Sprintf("CREATE (%v_%v:%v {arch:%q, name:{0}, node:{1}, package:{2}, timestamp:{3}, ip:{4}, region:{5}, zone:{6}})", archaius.Conf.Arch+ss, names.Instance(node), names.Service(node), archaius.Conf.Arch+ss))
	if err != nil {
		log.Fatal(err)
	}
	_, err = nodestmt.Exec(names.Instance(node), node, pack, tstamp, dhcp.Lookup(node), names.Region(node), names.Zone(node))
	if err != nil {
		log.Fatal(err)
	}
	nodestmt.Close()
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
	Write(fmt.Sprintf("MATCH (from:%v {name: %q}), (to:%v {name: %q})\nCREATE (from)-[:CONN {arch:%q, timestamp:%q}]->(to)", names.Service(source), names.Instance(source), names.Service(target), names.Instance(target), archaius.Conf.Arch+ss, tstamp))	
}

// record messages in neo4j as well as zipkin
func WriteFlow(source, target, call string, tnano int64, trace gotocol.TraceContextType) {
	if Enabled == false {
		return
	}
	if epoch == 0 {
		epoch = tnano
	}
	Write(fmt.Sprintf("MATCH (from:%v {name: %q}), (to:%v {name: %q})\nCREATE (from)-[:%v {arch:%q, timenano:%v, trace:%v}]->(to)", names.Service(source), names.Instance(source), names.Service(target), names.Instance(target), call, archaius.Conf.Arch+ss, tnano-epoch, trace))	
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
