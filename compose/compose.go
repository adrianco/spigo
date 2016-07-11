// Package compose reads docker-compose yaml files and generates architecture json output
// Would use https://github.com/docker/libcompose if it wasn't so mind-numbingly complicated
package compose

import (
	//"fmt"
	"github.com/adrianco/spigo/tooling/architecture"
	"github.com/cloudfoundry-incubator/candiedyaml"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
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
	file, err := os.Open(fn)
	if err != nil {
		log.Println("File does not exist:", err)
		return nil
	}
	defer file.Close()
	document := new(interface{})
	decoder := candiedyaml.NewDecoder(file)
	err = decoder.Decode(document)
	if err != nil {
		log.Println("Couldn't decode yaml:", err)
		return nil
	}
	c2 := new(ComposeV2Yaml)
	cs := make(ComposeServices)
	switch comp := (*document).(type) {
	case map[interface{}]interface{}:
		for label, section := range comp {
			switch label {
			case "version":
				switch version := section.(type) {
				case string:
					c2.Version = version
				default:
					log.Printf("Version not a string: %v", section)
					return nil
				}
			case "services":
				//fmt.Printf("Got services %v\n", section)
				switch services := section.(type) {
				case map[interface{}]interface{}:
					for name, options := range services {
						switch optionmap := options.(type) {
						case map[interface{}]interface{}:
							//fmt.Printf("Service %v:\n", name)
							ca := new(ComposeAttributes)
							for option, values := range optionmap {
								ok := true
								switch value := values.(type) {
								case []interface{}:
									//fmt.Printf("    slice %v:%v\n", option, value)
									sv := make([]string, len(value))
									for i, v := range value {
										sv[i], ok = v.(string)
										if !ok {
											log.Printf("Can't convert %v:%v to string\n", option, value)
										}
									}
									switch option {
									case "volumes":
										ca.Volumes = sv
									case "ports":
										ca.Ports = sv
									case "links":
										ca.Links = sv
									case "networks":
										ca.Networks = sv
									default:
										log.Printf("option ignored %v:%v\n", option, value)
									}
								case string:
									//fmt.Printf("    string %v:%v\n", option, value)
									switch option {
									case "build":
										ca.Build = value
									case "image":
										ca.Image = value
									default:
										log.Printf("option ignored %v:%v\n", option, value)
									}
								default:
									log.Printf("    not a string or slice %v:%v\n", option, value)
								}

							}
							cs[name.(string)] = *ca
						default:
							log.Printf("Couldn't find options in %v", optionmap)
						}
					}
				default:
					log.Printf("Couldn't find services in %v", services)
				}
				c2.Services = cs
			case "networks":
				switch networks := section.(type) {
				case map[interface{}]interface{}:
					c2.Networks = make(map[string]interface{})
					for name, options := range networks {
						//fmt.Printf("network %v:%v\n", name, options)
						s, ok := name.(string)
						if ok {
							c2.Networks[s] = options
						} else {
							log.Printf("Can't convert %v:%v to string\n", name, options)
						}
					}
				}
			case "volumes":
				switch volumes := section.(type) {
				case map[interface{}]interface{}:
					c2.Volumes = make(map[string]interface{})
					for name, options := range volumes {
						//fmt.Printf("volume %v:%v\n", name, options)
						s, ok := name.(string)
						if ok {
							c2.Volumes[s] = options
						} else {
							log.Printf("Can't convert %v:%v to string\n", name, options)
						}
					}
				}
			default:
				log.Printf("No matching section: %v\n", label)
			}
		}
	default:
		log.Println("Couldn't find sections in compose v2 file")
	}
	c2.Services = cs
	return c2
}

func ComposeArch(name string, c *ComposeV2Yaml) {
	a := architecture.MakeArch(name, "compose yaml")
	nets := make(map[string][]string)
	for n, v := range c.Services {
		//fmt.Println("Compose: ", n, v.Image, v.Build, v.Links)
		co := v.Image
		if co == "" {
			co = v.Build
		}
		var links []string // change db:redis into db
		for _, l := range v.Links {
			links = append(links, strings.Split(l, ":")[0])
		}
		var networks []string
		for _, name := range v.Networks {
			nets[name] = append(nets[name], n) // map of which services refer to a network
			networks = append(networks, name)  // list of networks this service refers to
		}
		var volumes []string
		for _, nv := range v.Volumes {
			name := strings.Split(nv, ":")[0] // get root volume name
			volumes = append(volumes, name)   // list of volumes this service refers to
		}
		if n == "db" {
			architecture.AddContainer(a, n, "machine", "instance", co, "process", "staash", 1, 1, volumes)
		} else {
			if n == "redis" {
				architecture.AddContainer(a, n, "machine", "instance", co, "process", "store", 1, 1, links)
			} else {
				architecture.AddContainer(a, n, "machine", "instance", co, "process", "monolith", 1, 3, links)
			}
		}
		external := false
		for _, port := range v.Ports {
			if len(strings.Split(port, ":")) >= 2 {
				external = true
			}
		}
		if external {
			var extlink []string
			extlink = append(extlink, n)
			architecture.AddContainer(a, "www-"+n, "external", "", "", "", "denominator", 0, 0, extlink)
		}
	}
	for n, _ := range c.Networks {
		architecture.AddContainer(a, n, "network", "", "", "", "elb", 1, 0, nets[n])
	}
	for n, _ := range c.Volumes {
		architecture.AddContainer(a, n, "volume", "", "", "", "store", 1, 0, nil)
	}
	architecture.WriteFile(a, name)
}
