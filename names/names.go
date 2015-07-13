// names creates and accesses the nanoservice naming hierarchy

package names

import (
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"strings"
)

type hier int

// offsets into the dot separated name
const (
	arch      hier = iota // netflixoss - architecture or AWS account level
	region                // us-east-1a - AWS region or equivalent
	zone                  // zoneA      - availability zone or datacenter
	machine               // ecs        - container orchestrator or physical machine name
	instance              // docker:0   - specific booted instance of service, or "docker:X"
	container             // homepage:1 - container name:id within instance
	process               // node:100   - process name/pid within container
	service               // homepage   - service type or application name
	gopackage             // karyon     - go package of code to implement service, like AMI or VM
)

// Split out a component of a name
func Splitter(name string, offset hier) string {
	s := strings.Split(name, ".")
	if len(s) > int(offset) {
		return s[offset]
	} else {
		return ""
	}
}

const (
	FilterDefault   = "*.*.*.*.*"
	FilterReduce    = "*.*.*.*.."
	FilterContainer = "*.*."
)

// Filter a name to take out components. "a.b.c" filter "*.*" returns "a"
func Filter(name, filter string) string {
	n := strings.Split(name, ".")
	nl := len(n) - 1
	f := strings.Split(filter, ".")
	fl := len(f) - 1
	if fl < 0 || nl < 0 || fl > nl {
		return name
	}
	fn := make([]string, len(n))
	l := len(fn)
	// from the end to the start
	for {
		if fl < 0 || f[fl] != "*" {
			l--
			fn[l] = n[nl]
		}
		fl--
		nl--
		if nl < 0 || l < 0 {
			break
		}
	}
	return strings.Join(fn[l:], ".")
}

// Filter a node
func FilterNode(node string) string {
	if archaius.Conf.Filter {
		return Filter(node, FilterReduce)
	} else {
		if Container(node) == "" {
			return Filter(node, FilterDefault)
		} else {
			return Filter(node, FilterContainer)
		}
	}
}

// Filter an edge - two space separated nodes
func FilterEdge(fromTo string) string {
	var source, target string
	fmt.Sscanf(fromTo, "%s%s", &source, &target) // two space delimited names
	return fmt.Sprintf("%v %v", FilterNode(source), FilterNode(target))
}

// Make a service name from components and an index
func Make(a, r, z, s, g string, i int) string {
	return MakeContainer(a, r, z, fmt.Sprintf("%v%v", s, i), "", "", "", s, g)
}

// Make a container name from components and an index
func MakeContainer(a, r, z, m, i, c, p, s, g string) string {
	return fmt.Sprintf("%v.%v.%v.%v.%v.%v.%v.%v.%v", a, r, z, m, i, c, p, s, g)
}

// Extract architecture from a name
func Arch(name string) string {
	return Splitter(name, arch)
}

// Extract region from a name
func Region(name string) string {
	return Splitter(name, region)
}

// Return the other regions given one
func OtherRegions(name string, rnames []string) []string {
	var nr []string
	regions := len(rnames)
	for i, r := range rnames {
		if Region(name) == r {
			for j := 1; j < regions; j++ {
				nr = append(nr, rnames[(i+j)%regions])
			}
		}
	}
	return nr
}

// Extract zone from a name
func Zone(name string) string {
	return Splitter(name, zone)
}

// Return the other two zones, given one
func OtherZones(name string, znames [3]string) [2]string {
	var nz [2]string
	for i, z := range znames {
		if Zone(name) == z {
			nz[0] = znames[(i+1)%3]
			nz[1] = znames[(i+2)%3]
		}
	}
	return nz
}

// Extract the region and zone together
func RegionZone(name string) string {
	return Splitter(name, region) + ". " + Splitter(name, zone)
}

// Extract the machine from a name
func Machine(name string) string {
	return Splitter(name, machine)
}

// Extract the instance from a name
func Instance(name string) string {
	return Splitter(name, instance)
}

// Extract the container from a name
func Container(name string) string {
	return Splitter(name, container)
}

// Extract the process from a name
func Process(name string) string {
	return Splitter(name, process)
}

// Extract the service from a name
func Service(name string) string {
	return Splitter(name, service)
}

// Extract the AMI from a name
func AMI(name string) string {
	return Splitter(name, gopackage)
}

// Extract the package (same as AMI) from a name
func Package(name string) string {
	return AMI(name)
}
