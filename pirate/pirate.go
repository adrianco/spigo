// The pirate package defines a participant in the network,
// which listens to the FSM and to other pirates and
// independently decides whether to make or break promises and behave.
package pirate

import (
	//	"fmt"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphml"
	"time"
)

// all configuration and state is sent via messages
func Listen(listener chan gotocol.Message) {
	dunbar := 10 // starting point for how many buddies to remember
	buddies := make(map[string]chan gotocol.Message, dunbar)
	var fsm chan gotocol.Message // remember how to talk back to creator
	var name string              // remember my name
	chatTicker := time.NewTicker(time.Hour)
	chatTicker.Stop()
	for {
		select {
		case msg := <-listener:
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
			case gotocol.Chat:
				// setup the ticker to run at the specified rate
				d, e := time.ParseDuration(msg.Intention)
				if e == nil && d >= time.Millisecond && d <= time.Hour {
					chatTicker = time.NewTicker(d)
				}
			case gotocol.Goodbye:
				fsm <- gotocol.Message{gotocol.Goodbye, nil, name}
				return
			}
		case <-chatTicker.C:
			// use Namedrop to tell the last buddy about the first
			var firstBuddyName string
			var firstBuddyChan, lastBuddyChan chan gotocol.Message
			if len(buddies) >= 2 {
				for name, ch := range buddies {
					if firstBuddyName == "" {
						firstBuddyName = name
						firstBuddyChan = ch
					} else {
						lastBuddyChan = ch
					}
					go gotocol.Send(lastBuddyChan, gotocol.Message{gotocol.NameDrop, firstBuddyChan, firstBuddyName})
				}
			}
		}
	}
}
