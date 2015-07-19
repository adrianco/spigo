// Package compose reads docker-compose yaml files and generates architecture json output
package compose

import (
	//"fmt"
	"github.com/adrianco/spigo/architecture"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strings"
)

// Compose Attribute maps to attributes of a microservice
type ComposeAttributes struct {
	Build string   `yaml:"build,omitempty"`
	Image string   `yaml:"image,omitempty"`
	Links []string `yaml:"links,omitempty"`
}

// Compose type to extract interesting data from compose yaml
type ComposeYaml map[string]ComposeAttributes

// ReadCompose
func ReadCompose(fn string) ComposeYaml {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	var c ComposeYaml
	e := yaml.Unmarshal(data, &c)
	if e == nil {
		return c
	} else {
		log.Fatal(e)
		return nil
	}
}

func ComposeArch(name string, c ComposeYaml) {
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
