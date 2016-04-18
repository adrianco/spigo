// tests for names
package names

import (
	"fmt"
	"github.com/adrianco/spigo/tooling/archaius"
	"testing"
)

func TestNames(t *testing.T) {
	name := Make("netflixoss", "us-east-1", "zoneA", "cassTurtle", "priamCassandra", 0)
	if name != "netflixoss.us-east-1.zoneA..cassTurtle00...cassTurtle.priamCassandra" {
		t.Fail()
	}
	fmt.Println("             " + name)
	fmt.Println("*:           " + Filter(name, "*"))
	fmt.Println("*.*:         " + Filter(name, "*.*"))
	fmt.Println("*.*.*:       " + Filter(name, "*.*.*"))
	fmt.Println("*.*.*.*:     " + Filter(name, "*.*.*.*"))
	fmt.Println("*..*.*.*.*:  " + Filter(name, "*..*.*.*.*"))
	fmt.Println("*.:          " + Filter(name, "*."))
	fmt.Println("*.*..:       " + Filter(name, "*.*.."))
	fmt.Println("FilterDefault " + FilterNode(name))
	archaius.Conf.Filter = true
	fmt.Println("FilterReduce " + FilterNode(name))
	archaius.Conf.Filter = false
	fmt.Println("FilterContainer " + FilterNode("container.us-east-1.zoneA.ecs:1.frontend:1.adrianco/node.node:1.homepage.karyon"))
	fmt.Println("Edge:        " + FilterEdge(fmt.Sprintf("%v %v", name, name)))
	fmt.Println("arch:        " + Arch(name))
	fmt.Println("region:      " + Region(name))
	fmt.Println("zone:        " + Zone(name))
	fmt.Println("service:     " + Service(name))
	fmt.Println("AMI/Package: " + AMI(name) + " " + Package(name))
	fmt.Println("Machine:     " + Machine(name))
	fmt.Println("Instance:    " + Instance(name))
	fmt.Println("Container:   " + Container(name))
	fmt.Println("Process:     " + Process(name))
	fmt.Printf("OtherZones:  %v\n", OtherZones(name, archaius.Conf.ZoneNames))
	fmt.Printf("OtherRegions:%v\n", OtherRegions(name, archaius.Conf.RegionNames[:]))
	fmt.Println("Pirates:     " + Package("fsm.atlantic.bermuda..blackbeard00...blackbeard.pirate"))
}
