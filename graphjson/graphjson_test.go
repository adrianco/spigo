// graphjson tests
package graphjson

import (
	"encoding/json"
	"fmt"
	"testing"
)

func try(t string) {
	v := new(GraphVersion)
	err := json.Unmarshal([]byte(t), v)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Version: ", v.Version)
	switch v.Version {
	case "spigo-0.3":
		fallthrough
	case "spigo-0.4":
		g := new(GraphV0r4)
		err = json.Unmarshal([]byte(t), g)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Architecture: ", g.Arch)
		newJSON, err := json.Marshal(g)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(newJSON))
	default:
		fmt.Println("Uknown version ", v.Version)
	}
}

// reader parses graphjson
func TestGraph(t *testing.T) {
	testJSONstringV0r3 := `
                {
                "arch":"fsm",
                "version":"spigo-0.3",
                "args":"[spigo -j -p=5 -d=0]",
                "graph":[
                        { "node":"Pirate1", "service":"pirate" },
                        { "node":"Pirate2", "service":"pirate" },
                        { "node":"Pirate3", "service":"pirate" },
                        { "node":"Pirate4", "service":"pirate" },
                        { "node":"Pirate5", "service":"pirate" },
                        { "edge":"e1", "source":"Pirate1", "target":"Pirate2" },
                        { "edge":"e2", "source":"Pirate1", "target":"Pirate2" },
                        { "edge":"e3", "source":"Pirate2", "target":"Pirate3" },
                        { "edge":"e4", "source":"Pirate2", "target":"Pirate4" },
                        { "edge":"e5", "source":"Pirate3", "target":"Pirate4" },
                        { "edge":"e6", "source":"Pirate3", "target":"Pirate4" },
                        { "edge":"e7", "source":"Pirate4", "target":"Pirate3" },
                        { "edge":"e8", "source":"Pirate4", "target":"Pirate1" },
                        { "edge":"e9", "source":"Pirate5", "target":"Pirate1" }
                        ]
                }`

	testJSONstringV0r4 := `
		{
		  "arch":"lamp",
		  "version":"spigo-0.4",
		  "args":"[./spigo -a lamp -d 1 -j -p 25]",
		  "date":"2015-04-26T23:52:45.959905585+12:00",
		  "graph":[
		    {"node":"lamp.us-east-1.zoneA.rds-mysql.store.rds-mysql0","package":"store","timestamp":"2015-04-26T23:52:45.960398393+12:00"},
		    {"node":"lamp.us-east-1.zoneB.rds-mysql.store.rds-mysql1","package":"store","timestamp":"2015-04-26T23:52:45.96054332+12:00"},
		    {"edge":"e1","source":"lamp.us-east-1.zoneB.rds-mysql.store.rds-mysql1","target":"lamp.us-east-1.zoneA.rds-mysql.store.rds-mysql0","timestamp":"2015-04-26T23:52:45.960569638+12:00"},
		    {"node":"lamp.us-east-1.zoneA.memcache.store.memcache0","package":"store","timestamp":"2015-04-26T23:52:45.960604405+12:00"},
		    {"node":"lamp.us-east-1.zoneA.phpweb.monolith.phpweb0","package":"monolith","timestamp":"2015-04-26T23:52:45.960622407+12:00"},
		    {"edge":"e2","source":"lamp.us-east-1.zoneA.phpweb.monolith.phpweb0","target":"lamp.us-east-1.zoneA.rds-mysql.store.rds-mysql0","timestamp":"2015-04-26T23:52:45.960638357+12:00"},
		    {"edge":"e3","source":"lamp.us-east-1.zoneA.phpweb.monolith.phpweb0","target":"lamp.us-east-1.zoneB.rds-mysql.store.rds-mysql1","timestamp":"2015-04-26T23:52:45.960652525+12:00"},
		    {"node":"lamp.us-east-1.zoneB.phpweb.monolith.phpweb1","package":"monolith","timestamp":"2015-04-26T23:52:45.960664447+12:00"},
		    {"edge":"e4","source":"lamp.us-east-1.zoneB.phpweb.monolith.phpweb1","target":"lamp.us-east-1.zoneA.rds-mysql.store.rds-mysql0","timestamp":"2015-04-26T23:52:45.960677597+12:00"},
		    {"edge":"e5","source":"lamp.us-east-1.zoneB.phpweb.monolith.phpweb1","target":"lamp.us-east-1.zoneB.rds-mysql.store.rds-mysql1","timestamp":"2015-04-26T23:52:45.96069075+12:00"},
		    {"node":"lamp.us-east-1.zoneC.phpweb.monolith.phpweb2","package":"monolith","timestamp":"2015-04-26T23:52:45.960702209+12:00"},
		    {"edge":"e6","source":"lamp.us-east-1.zoneC.phpweb.monolith.phpweb2","target":"lamp.us-east-1.zoneA.rds-mysql.store.rds-mysql0","timestamp":"2015-04-26T23:52:45.96071532+12:00"},
		    {"edge":"e7","source":"lamp.us-east-1.zoneC.phpweb.monolith.phpweb2","target":"lamp.us-east-1.zoneB.rds-mysql.store.rds-mysql1","timestamp":"2015-04-26T23:52:45.96072856+12:00"},
		    {"node":"lamp.us-east-1.zoneA.phpweb.monolith.phpweb3","package":"monolith","timestamp":"2015-04-26T23:52:45.96073986+12:00"},
		    {"edge":"e8","source":"lamp.us-east-1.zoneA.phpweb.monolith.phpweb3","target":"lamp.us-east-1.zoneA.rds-mysql.store.rds-mysql0","timestamp":"2015-04-26T23:52:45.960752752+12:00"},
		    {"edge":"e9","source":"lamp.us-east-1.zoneA.phpweb.monolith.phpweb3","target":"lamp.us-east-1.zoneB.rds-mysql.store.rds-mysql1","timestamp":"2015-04-26T23:52:45.960765849+12:00"},
		    {"node":"lamp.us-east-1.*.www-elb.elb.www-elb0","package":"elb","timestamp":"2015-04-26T23:52:45.960776833+12:00"},
		    {"edge":"e10","source":"lamp.us-east-1.*.www-elb.elb.www-elb0","target":"lamp.us-east-1.zoneA.phpweb.monolith.phpweb0","timestamp":"2015-04-26T23:52:45.960789708+12:00"},
		    {"edge":"e11","source":"lamp.us-east-1.*.www-elb.elb.www-elb0","target":"lamp.us-east-1.zoneB.phpweb.monolith.phpweb1","timestamp":"2015-04-26T23:52:45.96081064+12:00"},
		    {"edge":"e12","source":"lamp.us-east-1.*.www-elb.elb.www-elb0","target":"lamp.us-east-1.zoneA.phpweb.monolith.phpweb3","timestamp":"2015-04-26T23:52:45.960823404+12:00"},
		    {"edge":"e13","source":"lamp.us-east-1.*.www-elb.elb.www-elb0","target":"lamp.us-east-1.zoneC.phpweb.monolith.phpweb2","timestamp":"2015-04-26T23:52:45.960835814+12:00"},
		    {"node":"lamp.*.*.www.denominator.www0","package":"denominator","timestamp":"2015-04-26T23:52:45.9608483+12:00"},
		    {"edge":"e14","source":"lamp.*.*.www.denominator.www0","target":"lamp.us-east-1.*.www-elb.elb.www-elb0","timestamp":"2015-04-26T23:52:45.960860647+12:00"},
		    {"edge":"e15","source":"lamp.us-east-1.zoneA.rds-mysql.store.rds-mysql0","target":"lamp.us-east-1.zoneB.rds-mysql.store.rds-mysql1","timestamp":"2015-04-26T23:52:46.960271113+12:00"}
		  ]
		}`

	try(testJSONstringV0r3)
	try(testJSONstringV0r4)
}
