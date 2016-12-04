// Package archaius holds all configuration information, named after the netflixoss project
package archaius

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

// Configuration information for spigo
type Configuration struct {
	// Arch names the architecture pattern being simulated
	Arch string `json:"arch"`

	// GraphmlFile is set to a filename to turn on GraphML logging
	GraphmlFile string `json:"graphmlfile"`

	// GraphjsonFile is set to a filename to turn on GraphML logging
	GraphjsonFile string `json:"graphjsonfile"`

	// Neo4jURL is pointed at a database instance to turn on GraphML logging
	Neo4jURL string `json:"neo4jurl"`

	// RunDuration is the time in seconds to let the microservices chat
	RunDuration time.Duration `json:"runduration"`

	// Dunbar is a population scale factor
	Dunbar int `json:"dunbar"`

	// Population is the number of microservices in a network
	Population int `json:"population"`

	// Msglog if true, log each message received on the console
	Msglog bool `json:"msglog"`

	// Regions is the number of regions to create
	Regions int `json:"regions"`

	// RegionNames is the default names of the regions
	RegionNames []string `json:"regionnames"`

	// IPRanges maps an IP address range to each region and zone
	IPRanges [][]string `json:"ipranges"`

	// ZoneNames is the default names of the zones
	ZoneNames []string `json:"zonenames"`

	// Collect turns on Metrics collection
	Collect bool `json:"collect"`

	// Kafka turns on Zipkin compatible Flow export if array of host:port strings is not empty
	Kafka []string `json:"kafka"`

	// StopStep stops building new microservices at this step, 0 means don't stop
	StopStep int `json:"stopstep"`

	// EurekaPoll interval in seconds
	EurekaPoll string `json:"eurekapoll"`

	// Filter spec for output names to simplify graph
	Filter bool `json:"filter"`

	// Keys and values for configuring services, passed in as one string
	Keyvals string `json:"keyvals"`
}

// Conf data instance
var Conf = Configuration{
	RegionNames: []string{"us-east-1", "us-west-2", "eu-west-1", "eu-central-1", "ap-southeast-1", "ap-southeast-2"},
	ZoneNames:   []string{"zoneA", "zoneB", "zoneC"},
	IPRanges: [][]string{
		{"54.198.", "54.221.", "50.19."},  // Virginia us-east-1 actual AWS IP/16 ranges
		{"54.245.", "54.244.", "54.214."}, // Oregon us-west-2 actual AWS IP/16 ranges
		{"54.247.", "54.246.", "54.288."}, // Ireland eu-west-1 actual AWS IP/16 ranges
		{"54.93.", "54.28.", "54.78."},    // Frankfurt eu-central-1 actual AWS IP/16 ranges plus 54.78  stolen from Ireland
		{"54.251.", "54.254.", "54.255."}, // Singapore ap-southeast-1 actual AWS IP/16 ranges
		{"54.252.", "54.253.", "54.206."}, // Australia ap-southeast-2 actual AWS IP/16 ranges
	},
}

func init() {
	verifyConfig()
}

// verify the sizes of arrays above are equal at runtime
func verifyConfig() {
	if len(Conf.RegionNames) != len(Conf.IPRanges) {
		log.Fatal(fmt.Sprintf("RegionNames count (%d) does not match IPRanges count (%d)", len(Conf.RegionNames), len(Conf.IPRanges)))
	}
	for i := range Conf.IPRanges {
		if len(Conf.ZoneNames) != len(Conf.IPRanges[i]) {
			log.Fatal(fmt.Sprintf("ZoneNames count (%d) does not match IPRanges[%d] count (%d)", len(Conf.ZoneNames), i, len(Conf.IPRanges[i])))
		}
	}
}

// Key finds a value given a key
func Key(c Configuration, k string) string {
	if c.Keyvals == "" {
		return ""
	}
	kv := strings.Split(c.Keyvals, ":")
	if len(kv) == 2 && kv[0] == k {
		return kv[1]
	}
	return ""
}

// ReadConf parses json from a file
func ReadConf(config string) {
	fn := "json_arch/" + config + "_conf.json"
	log.Println("Loading config from " + fn)
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	FromJson(data)
	verifyConfig()
}

// WriteConf saves json to a file
func WriteConf() {
	fn := "json_arch/" + Conf.Arch + "_conf.json"
	log.Println("Saving config to " + fn)
	f, err := os.Create(fn)
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString(string(AsJson()))
	f.Close()
}

// AsJson returns current config as json
func AsJson() []byte {
	confJSON, _ := json.MarshalIndent(Conf, "", "    ")
	return confJSON
}

// FromJson imports a config from json
func FromJson(confJSON []byte) {
	json.Unmarshal(confJSON, &Conf)
}

// return formatted as string
func (Configuration) String() string {
	return fmt.Sprintf("Arch:       %v\nGraphML:    %v\nGraphJSON:  %v\nNeo4jURL:   %v\nRunDuration:%v\nDunbar:     %v\nPopulation: %v\nMsglog:     %v\nRegions:    %v\nRegionNames:%v\nZoneNames:  %v\nIPRanges:   %v\nCollect:    %v\nKafka:      %v\nStopStep:   %v\nEurekaPoll: %v\nKeyvals:    %v\n", Conf.Arch, Conf.GraphmlFile, Conf.GraphjsonFile, Conf.Neo4jURL, Conf.RunDuration, Conf.Dunbar, Conf.Population, Conf.Msglog, Conf.Regions, Conf.RegionNames, Conf.ZoneNames, Conf.IPRanges, Conf.Collect, Conf.Kafka, Conf.StopStep, Conf.EurekaPoll, Conf.Keyvals)
}
