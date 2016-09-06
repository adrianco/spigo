// architecture tests - just make sure the json conversions work
package architecture

import (
	"encoding/json"
	"fmt"
	"github.com/adrianco/spigo/tooling/archaius" // global configuration
	"testing"
	"time"
)

func try(t string) {
	a := new(archV0r1)
	err := json.Unmarshal([]byte(t), a)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Architecture: ", a.Arch)
	newJSON, err := json.Marshal(a)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(newJSON))
}

// reader parses graphjson
func TestGraph(t *testing.T) {
	testJSONarchV0r1 := `
                {
                "arch":"netflixoss",
                "version":"arch-0.1",
                "args":"[spigo -j -d=0 -a testContainer]",
		"victim":"homepage-node",
		"date":"2015-04-26T23:52:45.959905585+12:00",
                "services":[
		{ "name":"mysql", "package":"store", "regions":1, "count":2, "dependencies":[] },
		{ "name":"homepage", "machine":"ecs:4", "container":"adrianco/homepage-node", "package":"karyon", "regions":1, "count":9, "dependencies":["mysql"] },
		{ "name":"signup", "package":"karyon", "machine":"ecs:1", "container":"adrianco/signup-node", "process":"signup-node", "regions":1, "count":3, "dependencies":["mysql"] },
		{ "name":"signup-waf", "package":"karyon", "machine":"ecs:1", "container":"adrianco/signup-node", "process":"waf", "regions":1, "count":3, "dependencies":["signup-node"] },
		{ "name":"www-proxy", "package":"zuul", "regions":1, "count":3, "dependencies":["signup-waf", "homepage"] },
		{ "name":"www-elb", "package":"elb", "regions":1, "count":0, "dependencies":["www-proxy"] },
		{ "name":"www", "package":"denominator", "regions":0, "count":0, "dependencies":["www-elb"] }
                ]
                }`

	archaius.Conf.Arch = "test"
	//archaius.Conf.GraphmlFile = ""
	//archaius.Conf.GraphjsonFile = ""
	archaius.Conf.RunDuration = 2 * time.Second
	archaius.Conf.Dunbar = 50
	archaius.Conf.Population = 50
	//archaius.Conf.Msglog = false
	archaius.Conf.Regions = 1
	//archaius.Conf.Collect = false
	//archaius.Conf.StopStep = 0
	archaius.Conf.EurekaPoll = "1s"
	try(testJSONarchV0r1)
	//ReadArch("testDuplicate") // these three are designed to fail, uncomment one at a time to check
	//ReadArch("testMissingDep")
	//ReadArch("testBadPackage")
	a := ReadArch("test")
	fmt.Println(a)
	var services []string
	var deps []Connection
	ListDependencies(a, &services, &deps)
	fmt.Println("Services and Dependencies")
	fmt.Println(services)
	fmt.Println(deps)
	Start(a)
}
