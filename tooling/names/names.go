// Package names creates and accesses the nanoservice naming hierarchy

package names

import (
	"fmt"
	"github.com/adrianco/spigo/tooling/archaius"
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

// Splitter for a component of a name
func Splitter(name string, offset hier) string {
	s := strings.Split(name, ".")
	if len(s) > int(offset) {
		return s[offset]
	}
	return ""
}

const (
	FilterDefault   = "*..*.*.*.*"
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

// FilterNode to simplify its name
func FilterNode(node string) string {
	if archaius.Conf.Filter {
		return Filter(node, FilterReduce)
	}
	if Container(node) == "" {
		return Filter(node, FilterDefault)
	}
	return Filter(node, FilterContainer)
}

// FilterEdge - two space separated nodes
func FilterEdge(fromTo string) string {
	var source, target string
	fmt.Sscanf(fromTo, "%s%s", &source, &target) // two space delimited names
	return fmt.Sprintf("%v %v", FilterNode(source), FilterNode(target))
}

// Make a service name from components and an index
func Make(a, r, z, s, g string, i int) string {
	return MakeContainer(a, r, z, "", fmt.Sprintf("%v%02v", s, i), "", "", s, g)
}

// MakeContainer name from components and an index
func MakeContainer(a, r, z, m, i, c, p, s, g string) string {
	return fmt.Sprintf("%v.%v.%v.%v.%v.%v.%v.%v.%v", a, r, z, m, i, c, p, s, g)
}

// Arch extract from a name
func Arch(name string) string {
	return Splitter(name, arch)
}

// Region extract from a name
func Region(name string) string {
	return Splitter(name, region)
}

// OtherRegions given one
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

// Zone extract from a name
func Zone(name string) string {
	return Splitter(name, zone)
}

// OtherZones, given one
func OtherZones(name string, znames []string) []string {
	var nz []string
	for _, z := range znames {
		if Zone(name) != z {
			nz = append(nz, z)
		}
	}
	return nz
}

// RegionZone together
func RegionZone(name string) string {
	return Splitter(name, region) + ". " + Splitter(name, zone)
}

// Machine from a name
func Machine(name string) string {
	return Splitter(name, machine)
}

// Instance from a name
func Instance(name string) string {
	return Splitter(name, instance)
}

// Container from a name
func Container(name string) string {
	return Splitter(name, container)
}

// Process from a name
func Process(name string) string {
	return Splitter(name, process)
}

// Service from a name
func Service(name string) string {
	return Splitter(name, service)
}

// AMI from a name
func AMI(name string) string {
	return Splitter(name, gopackage)
}

// Package (same as AMI) from a name
func Package(name string) string {
	return AMI(name)
}
