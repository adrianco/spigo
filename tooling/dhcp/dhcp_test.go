// test dhcp
package dhcp

import (
	"fmt"
	. "github.com/adrianco/spigo/tooling/archaius"
	"github.com/adrianco/spigo/tooling/names"
	"testing"
	"time"
)

func TestConf(t *testing.T) {
	Conf.Arch = "testarch"
	Conf.GraphmlFile = "graphml"
	Conf.GraphjsonFile = "graphjson"
	Conf.RunDuration = 10 * time.Second
	Conf.Dunbar = 100
	Conf.Population = 100
	Conf.Msglog = true
	Conf.Regions = 2
	Conf.Collect = true
	Conf.StopStep = 2
	Conf.EurekaPoll = "1s"
	for _, r := range Conf.RegionNames {
		for i, z := range Conf.ZoneNames {
			// make two different names to make sure the IP is unique and make sure it is stored
			n1 := names.Make(Conf.Arch, r, z, "lookup", "dhcp", i)
			fmt.Println(n1, Lookup(n1))
			n2 := names.Make(Conf.Arch, r, z, "lookup", "dhcp", 100+i)
			fmt.Println(n2, Lookup(n2))
			fmt.Println(n2, Lookup(n2)) // check stored lookup returns the same thing
		}
	}
}
