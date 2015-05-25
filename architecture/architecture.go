// Package architecture reads a microservice architecture definition from a file
// It creates and controls a collection of aws and netflix application microservices
package architecture

import (
	"encoding/json"
	"github.com/adrianco/spigo/archaius" // global configuration
	"github.com/adrianco/spigo/asgard"   // tools to create an architecture
	"io/ioutil"
	"log"
)

type archV0r0 struct {
	Arch        string        `json:"arch"`
	Version     string        `json:"version"`
	Description string        `json:"description,omitempty"`
	Args        string        `json:"args,omitempty"`
	Date        string        `json:"date,omitempty"`
	Victim      string        `json:"victim,omitempty"`
	Services    []serviceV0r0 `json:"services"`
}

type serviceV0r0 struct {
	Name         string   `json:"name"`
	Package      string   `json:"package"`
	Regions      int      `json:"regions,omitempty"`
	Count        int      `json:"count"`
	Dependencies []string `json:"dependencies"`
}

// Start architecture
func Start(a *archV0r0) {
	var r string
	if archaius.Conf.Population < 1 {
		log.Fatal("architecture: can't create less than 1 microservice")
	} else {
		log.Printf("architecture: scaling to %v%%", archaius.Conf.Population)
	}
	asgard.CreateChannels()
	asgard.CreateEureka() // service registries for each zone

	for _, s := range a.Services {
		log.Printf("Starting: %v\n", s)
		r = asgard.Create(s.Name, s.Package, s.Regions*archaius.Conf.Regions, s.Count*archaius.Conf.Population/100, s.Dependencies...)
	}
	asgard.Run(r, a.Victim) // run the last service in the list, and point chaos monkey at the victim
}

// ReadArch parses archjson
func ReadArch(arch string) *archV0r0 {
	fn := "json_arch/" + arch + "_arch.json"
	log.Println("Loading architecture from " + fn)
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	a := new(archV0r0)
	e := json.Unmarshal(data, a)
	if e == nil {
		names := make(map[string]bool, 10)
		names[asgard.EurekaPkg] = true // special case to allow cross region references
		packs := make(map[string]bool, 10)
		for _, p := range asgard.Packages {
			packs[p] = true
		}
		// map all the service names and check packages exist
		for _, s := range a.Services {
			if names[s.Name] == true {
				log.Println(names)
				log.Println(s)
				log.Fatal("Duplicate service name in architecture: " + s.Name)
			} else {
				names[s.Name] = true
			}
			if packs[s.Package] != true {
				log.Println(packs)
				log.Println(s)
				log.Fatal("Unknown package name in architecture: " + s.Package)
			}
		}
		// check all the dependencies
		for _, s := range a.Services {
			for _, d := range s.Dependencies {
				if names[d] == false {
					log.Println(names)
					log.Println(s)
					log.Fatal("Unknown dependency name in architecture: " + d)
				}
			}
		}
		log.Printf("Architecture: %v %v\n", a.Arch, a.Description)
		return a
	} else {
		log.Fatal(e)
		return nil
	}
}
