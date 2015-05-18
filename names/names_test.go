// tests for names
package names

import (
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"testing"
)

func TestNames(t *testing.T) {
	name := Make("netflixoss", "us-east-1", "zoneA", "cassTurtle", "priamCassandra", 0)
	if name != "netflixoss.us-east-1.zoneA.cassTurtle.priamCassandra.cassTurtle0" {
		t.Fail()
	}
	fmt.Println(name)
	fmt.Println("arch:        " + Arch(name))
	fmt.Println("region:      " + Region(name))
	fmt.Println("zone:        " + Zone(name))
	fmt.Println("service:     " + Service(name))
	fmt.Println("AMI/Package: " + AMI(name) + " " + Package(name))
	fmt.Println("Instance:    " + Instance(name))
	fmt.Printf("OtherZones:  %v\n", OtherZones(name, archaius.Conf.ZoneNames))
	fmt.Printf("OtherRegions:%v\n", OtherRegions(name, archaius.Conf.RegionNames[:]))
	fmt.Println("Pirates:     " + Package("fsm.atlantic.bermuda.blackbeard.pirate.blackbeard0"))
}
