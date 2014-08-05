// participant in the network, listens to the FSM and to other pirates
// independently decides whether to make or break promises and behave

package pirate

import (
//	"fmt"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphml"
)

// all configuration and state is sent via messages
func Listen(listener chan gotocol.Message) {
	dunbar := 10 // starting point for how many buddies to remember
	buddies := make(map[string]chan gotocol.Message, dunbar)
	var fsm chan gotocol.Message // remember how to talk back to creator
	var name string              // remember my name
	var msg gotocol.Message
	for {
		msg = <-listener
		//fmt.Println(msg)
		switch msg.Imposition {
		case gotocol.Hello:
			switch {
			case name == "":
				// if I don't have a name yet
				fsm = msg.ResponseChan // remember who named me
				name = msg.Intention
			}
		case gotocol.NameDrop:
			// don't remember too many buddies and don't talk to myself
			if len(buddies) < dunbar && msg.Intention != name {
				// remember how to talk to this buddy
				buddies[msg.Intention] = msg.ResponseChan
				graphml.Edge(msg.Intention, name)
			}
		case gotocol.Goodbye:
			// if my creator told me to die, reply
			//if msg.ResponseChan == fsm {
				fsm <- gotocol.Message{gotocol.Goodbye, nil, name}
				return
			//}
		}
	}
}
