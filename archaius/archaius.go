// package archaius holds all configuration information, named after the netflixoss project
package archaius

import (
	"encoding/json"
	"fmt"
	"time"
)

type Configuration struct {
	// Arch names the architecture pattern being simulated
	Arch string `json:"arch,omitempty"`

	// GraphmlFile is set to a filename to turn on GraphML logging
	GraphmlFile string `json:"graphmlfile,omitempty"`

	// GraphjsonFile is set to a filename to turn on GraphML logging
	GraphjsonFile string `json:"graphjsonfile,omitempty"`

	// RunDuration is the time in seconds to let the microservices chat
	RunDuration time.Duration `json:"runduration,omitempty"`

	// Dunbar is a population scale factor
	Dunbar int `json:"dunbar,omitempty"`

	// Population is the number of microservices in a network
	Population int `json:"population,omitempty"`

	// Msglog if true, log each message received on the console
	Msglog bool `json:"msglog",omitempty"`

	// Regions is the number of regions to create
	Regions int `json:"regions,omitempty"`

	// RegionNames is the default names of the regions
	RegionNames [10]string `json:"regionnames,omitempty"`

	// Collect turns on Metrics collection
	Collect bool `json:"collect,omitempty"`

	// StopStep stops building new microservices at this step, 0 means don't stop
	StopStep int `json:"stopstep,omitempty"`

	// EurekaPoll interval in seconds
	EurekaPoll string `json:"eurekapoll, omitempty"`
}

var Conf Configuration

// return current config as json
func AsJson() []byte {
	confJSON, _ := json.Marshal(Conf)
	return confJSON
}

func FromJson(confJSON []byte) {
	json.Unmarshal(confJSON, &Conf)
}

// return formatted as string
func (Configuration) String() string {
	return fmt.Sprintf("Arch:       %v\nGraphML:    %v\nGraphJSON:  %v\nRunDuration:%v\nDunbar:     %v\nPopulation: %v\nMsglog:     %v\nRegions:    %v\nRegionNames:%v\nCollect:    %v\nStopStep:   %v\nEurekaPoll: %v\n", Conf.Arch, Conf.GraphmlFile, Conf.GraphjsonFile, Conf.RunDuration, Conf.Dunbar, Conf.Population, Conf.Msglog, Conf.Regions, Conf.RegionNames, Conf.Collect, Conf.StopStep, Conf.EurekaPoll)
}
