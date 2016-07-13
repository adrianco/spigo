// Package dhcp provides individual IP addresses by name, simulating dhcp
package dhcp

import (
	"fmt"
	"github.com/adrianco/spigo/tooling/archaius"
	"github.com/adrianco/spigo/tooling/names"
)

var (
	allocated [][]int
	mapped    map[string]string
)

func init() {
	allocated = make([][]int, len(archaius.Conf.RegionNames))
	for i := range allocated {
		allocated[i] = make([]int, len(archaius.Conf.ZoneNames))
	}
	mapped = make(map[string]string, archaius.Conf.Population)
}

// Lookup a name and return the IP address for that name
func Lookup(name string) string {
	ip := mapped[name]
	if ip != "" {
		return ip
	}
	// find indexes for matching zone and region in the config
	r := names.Region(name)
	ri := 0
	z := names.Zone(name)
	zi := 0
	for i, rr := range archaius.Conf.RegionNames {
		if rr == r {
			ri = i
			break
		}
	}
	for i, zr := range archaius.Conf.ZoneNames {
		if zr == z {
			zi = i
			break
		}
	}
	// increment first to avoid IP 0.0 and get the node counter in the region/zone
	allocated[ri][zi]++
	node := allocated[ri][zi]
	// format as xxx.xxx.xxx.xxx
	addr := fmt.Sprintf("%v%v.%v", archaius.Conf.IPRanges[ri][zi], node/256, node%256)
	mapped[name] = addr
	return addr
}
