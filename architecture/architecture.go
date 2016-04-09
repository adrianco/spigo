// Package architecture reads a microservice architecture definition from a file
// It creates and controls a collection of aws and netflix application microservices
package architecture

import (
	"encoding/json"
	"fmt"
	"github.com/adrianco/spigo/archaius"     // global configuration
	"github.com/adrianco/spigo/asgard"       // tools to create an architecture
	"github.com/adrianco/spigo/packagenames" // name definitions
	"io/ioutil"
	"log"
	"os"
	"time"
)

type archV0r1 struct {
	Arch        string          `json:"arch"`
	Version     string          `json:"version"`
	Description string          `json:"description,omitempty"`
	Args        string          `json:"args,omitempty"`
	Date        string          `json:"date,omitempty"`
	Victim      string          `json:"victim,omitempty"`
	Services    []containerV0r0 `json:"services"`
}

type serviceV0r0 struct {
	Name         string   `json:"name"`
	Package      string   `json:"package"`
	Regions      int      `json:"regions,omitempty"`
	Count        int      `json:"count"`
	Dependencies []string `json:"dependencies"`
}

type containerV0r0 struct {
	Name         string   `json:"name"`
	Machine      string   `json:"machine,omitempty"`
	Instance     string   `json:"instance,omitempty"`
	Container    string   `json:"container,omitempty"`
	Process      string   `json:"process,omitempty"`
	Gopackage    string   `json:"package"`
	Regions      int      `json:"regions,omitempty"`
	Count        int      `json:"count"`
	Dependencies []string `json:"dependencies"`
}

// Start architecture
func Start(a *archV0r1) {
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
		r = asgard.Create(s.Name, s.Gopackage, s.Regions*archaius.Conf.Regions, s.Count*archaius.Conf.Population/100, s.Dependencies...)
	}
	asgard.Run(r, a.Victim) // run the last service in the list, and point chaos monkey at the victim
}

// ReadArch parses archjson
func ReadArch(arch string) *archV0r1 {
	fn := "json_arch/" + arch + "_arch.json"
	log.Println("Loading architecture from " + fn)
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	a := new(archV0r1)
	e := json.Unmarshal(data, a)
	if e == nil {
		names := make(map[string]bool)
		names[packagenames.EurekaPkg] = true // special case to allow cross region references
		packs := make(map[string]bool)
		for _, p := range packagenames.Packages {
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
			if packs[s.Gopackage] != true {
				log.Println(packs)
				log.Println(s)
				log.Fatal("Unknown package name in architecture: " + s.Gopackage)
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

// Make a new architecture object
func MakeArch(arch, des string) *archV0r1 {
	a := new(archV0r1)
	a.Arch = arch
	a.Version = "arch-0.1"
	a.Description = des
	a.Args = fmt.Sprintf("%v", os.Args)
	a.Date = time.Now().Format(time.RFC3339Nano)
	a.Victim = ""
	return a
}

func AddContainer(a *archV0r1, name, machine, instance, container, process, gopackage string, regions, count int, dependencies []string) {
	var c containerV0r0
	c.Name = name
	c.Machine = machine
	c.Instance = instance
	c.Container = container
	c.Process = process
	c.Gopackage = gopackage
	c.Regions = regions
	c.Count = count
	c.Dependencies = dependencies
	a.Services = append(a.Services, c)
}

func Write(a *archV0r1) {
	b, err := json.Marshal(a)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		os.Stdout.Write(b)
	}
}
