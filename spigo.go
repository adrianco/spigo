// simulate protocol interactions in go - spigo
// terminology is a mix of promise theory and flying spaghetti monster lore

package main

import (
	"fmt"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/fsm"
	"github.com/adrianco/spigo/pirate"
)

func main() {
	fmt.Println("Spigo")
	const population = 10
	noodles := make(map[string]chan gotocol.Message, population)
	for i := 1; i <= population; i++ {
		name := fmt.Sprintf("Pirate%d", i)
		noodles[name] = make(chan gotocol.Message)
		go pirate.Listen(noodles[name])
	}
	fsm.Touch(noodles)
}
