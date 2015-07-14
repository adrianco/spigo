// Package compose reads docker-compose yaml files and generates architecture json output
package compose

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

// ReadCompose
func ReadCompose(compose string) *yaml.MapSlice {
	fn := "compose_yaml/" + compose + ".yml"
	log.Println("Loading compose yaml from " + fn)
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	c := new(yaml.MapSlice)
	e := yaml.Unmarshal(data, c)
	if e == nil {
		return c
	} else {
		log.Fatal(e)
		return nil
	}
}
