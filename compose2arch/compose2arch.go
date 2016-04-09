// utility to read a docker compose yaml file and write out an arch_json
package main

import (
	"flag"
	"github.com/adrianco/spigo/compose"
)

func main() {
	var fn string
	var v1 bool
	flag.StringVar(&fn, "file", "", "docker compose format yaml file")
	flag.BoolVar(&v1, "v1", false, "read from compose v1 format yaml file - default is v2")
	flag.Parse()
	if fn != "" {
		if v1 {
			c1 := compose.ReadCompose(fn)
			if c1 != nil {
				c2 := new(compose.ComposeV2Yaml)
				c2.Services = c1
				compose.ComposeArch(fn, c2)
			}
		} else {
			c2 := compose.ReadComposeV2(fn)
			if c2 != nil {
				compose.ComposeArch(fn, c2)
			}
		}
	} else {
		flag.PrintDefaults()
	}
}
