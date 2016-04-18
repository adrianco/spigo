// compose tests - just make sure the yaml conversions work
package compose

import (
	//"encoding/json"
	"fmt"
	"github.com/adrianco/spigo/tooling/archaius" // global configuration
	"github.com/adrianco/spigo/tooling/architecture"
	"github.com/cloudfoundry-incubator/candiedyaml"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func try(t string) {
	var c ComposeServices
	err := yaml.Unmarshal([]byte(t), &c)
	if err != nil {
		fmt.Println(err)
	}
	a := architecture.MakeArch("test", "compose yaml")
	for n, v := range c {
		fmt.Println("Compose: ", n, v.Image, v.Build, v.Links)
		c := v.Image
		if c == "" {
			c = v.Build
		}
		var links []string // change db:redis into db
		for _, l := range v.Links {
			links = append(links, strings.Split(l, ":")[0])
		}
		architecture.AddContainer(a, n, "machine", "instance", c, "process", "monolith", 1, 3, links)
	}
	fmt.Println(*a)
}

// test based on https://github.com/b00giZm/docker-compose-nodejs-examples/blob/master/05-nginx-express-redis-nodemon/docker-compose.yml
func TestGraph(t *testing.T) {
	testyaml := `
web:
  build: ./app
  volumes:
    - "app:/src/app"
  expose:
    - "3000"
  links:
    - "db:redis"
  command: nodemon -L app/bin/www

nginx:
  restart: always
  build: ./nginx/
  ports:
    - "80:80"
  volumes:
    - /www/public
  volumes_from:
    - web
  links:
    - web:web

db:
  image: redis
`

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
	fmt.Println("\nTesting Parser from Docker Compose V1 string")
	try(testyaml)

	fmt.Println("\nTesting file conversion from Docker Compose V1 compose_yaml/test.yml")
	c := ReadCompose("compose_yaml/test.yml")
	fmt.Println(c)

	fmt.Println("\nTesting Docker Compose V2 format input from compose_yaml/testV2.yml")
	file, err := os.Open("compose_yaml/testV2.yml")
	if err != nil {
		println("File does not exist:", err.Error())
		os.Exit(1)
	}
	defer file.Close()
	document := new(interface{})
	// too hard to parse V2 yaml with decoder, so use candiedyaml to walk the structure
	decoder := candiedyaml.NewDecoder(file)
	err = decoder.Decode(document)
	if err != nil {
		log.Fatal(err)
	}
	println("parsed yml:")
	cs := make(ComposeServices)
	switch comp := (*document).(type) {
	case map[interface{}]interface{}:
		for label, section := range comp {
			switch label {
			case "version":
				switch version := section.(type) {
				case string:
					fmt.Printf("Got version %v\n", version)
				default:
					fmt.Printf("Version not a string %v\n", section)
				}
			case "services":
				//fmt.Printf("Got services %v\n", section)
				switch services := section.(type) {
				case map[interface{}]interface{}:
					for name, options := range services {
						switch optionmap := options.(type) {
						case map[interface{}]interface{}:
							fmt.Printf("Service %v:\n", name)
							ca := new(ComposeAttributes)
							for option, values := range optionmap {
								ok := true
								switch value := values.(type) {
								case []interface{}:
									fmt.Printf("    slice %v:%v\n", option, value)
									sv := make([]string, len(value))
									for i, v := range value {
										sv[i], ok = v.(string)
										if !ok {
											fmt.Printf("not ok %v:%v\n", option, value)
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
										fmt.Printf("option ignored %v:%v\n", option, value)
									}
								case string:
									fmt.Printf("    string %v:%v\n", option, value)
									switch option {
									case "build":
										ca.Build = value
									case "image":
										ca.Image = value
									default:
										fmt.Printf("option ignored %v:%v\n", option, value)
									}
								default:
									fmt.Printf("    no match %v:%v\n", option, value)
								}

							}
							cs[name.(string)] = *ca
						default:
							fmt.Printf("Couldn't find services in %v", services)
						}
					}
				default:
					fmt.Println("Couldn't find services")
				}
				for k, v := range cs {
					fmt.Printf("Service %v:%v\n", k, v)
				}
			case "networks":
				switch networks := section.(type) {
				case map[interface{}]interface{}:
					for name, options := range networks {
						fmt.Printf("network %v:%v\n", name, options)
					}
				}
			case "volumes":
				switch volumes := section.(type) {
				case map[interface{}]interface{}:
					for name, options := range volumes {
						fmt.Printf("volume %v:%v\n", name, options)
					}
				}
			default:
				fmt.Printf("No matching section: %v\n", label)
			}
		}
	default:
		fmt.Println("Couldn't find sections in compose v2 file")
	}

	fmt.Println("\nTesting ReadComposeV2(compose_yaml/testV2.yml)")
	c2 := ReadComposeV2("compose_yaml/testV2.yml")
	fmt.Println(*c2)
	ComposeArch("composeTestResult", c2)
}
