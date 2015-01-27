// graphjson tests
package graphjson

import (
	"encoding/json"
	"fmt"
	"testing"
)

// reader parses graphjson
func TestGraph(t *testing.T) {
	test_json_string := `
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

	v := new(Graph_Version)
	json.Unmarshal([]byte(test_json_string), v)
	fmt.Println("Version: ", v.Version)
	switch v.Version {
	case "spigo-0.3":
		g := new(Graph_0_3)
		json.Unmarshal([]byte(test_json_string), g)
		fmt.Println("Architecture: ", g.Arch)
		new_json, _ := json.Marshal(g)
		fmt.Println(string(new_json))
	default:
		fmt.Println("Uknown version ", v.Version)
	}
}
