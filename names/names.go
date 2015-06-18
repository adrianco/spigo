// names creates and accesses the nanoservice naming hierarchy

package names

import (
	"fmt"
	"strings"
)

type hier int

// offsets into the dot separated name
const (
	arch      hier = iota // netflixoss - architecture or AWS account level
	region                // us-east-1a - AWS region or equivalent
	zone                  // zoneA      - AWS availability zone or equivalent
	service               // cassTurtle - service type or application name
	ami                   // priamCassandra - versioned package of code to implement service, like AMI
	instance              // cassTurtle0 - specific instance of service, like EC2 instance
	container             // docker container
	process               // process id within container
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

// Make a service name from components and an index
func Make(a, r, z, s, p string, i int) string {
	return fmt.Sprintf("%v.%v.%v.%v.%v.%v%v", a, r, z, s, p, s, i)
}

// Make a container name from components and an index
func MakeContainer(a, r, z, s, p, c string, i int) string {
	return fmt.Sprintf("%v.%v.%v.%v.%v.%v%v.%v.%v", a, r, z, s, p, s, i, c, i)
}

// Extract architecture from a name
func Arch(name string) string {
	return Splitter(name, arch)
}

// Extract region from a name
func Region(name string) string {
	return Splitter(name, region)
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

// Extract the region and zone together
func RegionZone(name string) string {
	return Splitter(name, region) + ". " + Splitter(name, zone)
}

// Extract the service from a name
func Service(name string) string {
	return Splitter(name, service)
}

// Extract the AMI from a name
func AMI(name string) string {
	return Splitter(name, ami)
}

// Extract the package (same as AMI) from a name
func Package(name string) string {
	return AMI(name)
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
