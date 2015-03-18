// names creates and accesses the nanoservice naming hierarchy

package names

import (
	"fmt"
	"strings"
)

type hier int

// offsets into the dot separated name
const (
	arch     hier = iota // netflixoss - architecture or AWS account level
	region               // us-east-1a - AWS region or equivalent
	zone                 // zoneA      - AWS availability zone or equivalent
	service              // cassTurtle - service type or application name
	ami                  // priamCassandra - versioned package of code to implement service, like AMI
	instance             // cassTurtle0 - specific instance of service, like EC2 instance
)

func Make(a, r, z, s, p string, i int) string {
	return fmt.Sprintf("%v.%v.%v.%v.%v.%v%v", a, r, z, s, p, s, i)
}

func Arch(name string) string {
	return strings.Split(name, ".")[arch]
}

func Region(name string) string {
	return strings.Split(name, ".")[region]
}

func Zone(name string) string {
	return strings.Split(name, ".")[zone]
}

func RegionZone(name string) string {
	s := strings.Split(name, ".")
	return s[region]+"."+s[zone]
}

func Service(name string) string {
	return strings.Split(name, ".")[service]
}

func AMI(name string) string {
	return strings.Split(name, ".")[ami]
}

func Package(name string) string {
	return AMI(name)
}

func Instance(name string) string {
	return strings.Split(name, ".")[instance]
}
