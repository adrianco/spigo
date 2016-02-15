// Package compose reads docker-compose yaml files and generates architecture json output
// Would use https://github.com/docker/libcompose if it wasn't so mind-numbingly complicated
package compose

import (
	//"fmt"
	"github.com/adrianco/spigo/architecture"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strings"
)

// Compose V1 Attribute maps to attributes of a microservice
type ComposeAttributes struct {
	Build    string   `yaml:"build,omitempty"`
	Image    string   `yaml:"image,omitempty"`
	Links    []string `yaml:"links,omitempty"`
	Volumes  []string `yaml:"volumes,omitempty"`
	Ports    []string `yaml:"ports,omitempty"`
	Networks []string `yaml:"networks,omitempty"`
}

// Compose type to extract interesting data from compose yaml version 1 file
type ComposeServices map[string]ComposeAttributes

// Compose type to extract interesting data from compose yaml version 2 file
type ComposeV2Yaml struct {
	Version  string                 `yaml:"2,omitempty"`
	Services ComposeServices        `yaml:"services,omitempty"`
	Networks map[string]interface{} `yaml:"networks,omitempty"`
	Volumes  map[string]interface{} `yaml:"volumes,omitempty"`
}

// ReadCompose
func ReadCompose(fn string) ComposeServices {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	var c ComposeServices
	e := yaml.Unmarshal(data, &c)
	if e == nil {
		return c
	} else {
		log.Println(e)
		return nil
	}
}

// ReadCompose for V2
func ReadComposeV2(fn string) *ComposeV2Yaml {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	var c ComposeV2Yaml
	e := yaml.Unmarshal(data, &c)
	if e == nil {
		return &c
	} else {
		log.Println(e)
		return nil
	}
}

func ComposeArch(name string, c ComposeServices) {
	a := architecture.MakeArch(name, "compose yaml")
	for n, v := range c {
		//fmt.Println("Compose: ", n, v.Image, v.Build, v.Links)
		co := v.Image
		if co == "" {
			co = v.Build
		}
		var links []string // change db:redis into db
		for _, l := range v.Links {
			links = append(links, strings.Split(l, ":")[0])
		}
		architecture.AddContainer(a, n, "machine", "instance", co, "process", "monolith", 1, 3, links)
	}
	architecture.Write(a)
}
