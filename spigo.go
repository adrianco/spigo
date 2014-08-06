// simulate protocol interactions in go - spigo
// terminology is a mix of promise theory and flying spaghetti monster lore

package main

import (
	"fmt"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/fsm"
	"github.com/adrianco/spigo/pirate"
	"github.com/adrianco/spigo/graphml"
)

func main() {
	const population = 100
	fmt.Println("Spigo population", population, "pirates")
	graphml.Setup()
	noodles := make(map[string]chan gotocol.Message, population)
	for i := 1; i <= population; i++ {
		name := fmt.Sprintf("Pirate%d", i)
		graphml.Node(name)
		noodles[name] = make(chan gotocol.Message)
		go pirate.Listen(noodles[name])
	}
	fsm.Touch(noodles)
	graphml.Close()
}
