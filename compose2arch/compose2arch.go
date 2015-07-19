// utility to read a docker compose yaml file and write out an arch_json
package main

import (
	"flag"
	"github.com/adrianco/spigo/compose"
)

func main() {
	var fn string
	flag.StringVar(&fn, "file", "", "docker compose format yaml file")
	flag.Parse()
	if fn != "" {
		compose.ComposeArch(fn, compose.ReadCompose(fn))
	} else {
		flag.PrintDefaults()
	}
}
