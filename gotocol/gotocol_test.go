// Tests for gotocol

package gotocol

import (
	"fmt"
	"testing"
)

// all configuration and state is sent via messages
func pirateListen(listener chan Message) {
	var buddy chan Message
	var msg Message
	for {
		msg = <-listener
		fmt.Println(msg)
		if msg.ResponseChan != nil {
			buddy = msg.ResponseChan
		}
		switch msg.Imposition {
		case Hello:
		case NameDrop:
			if buddy != nil {
				Message{Hello, listener, "Pirate"}.GoSend(buddy)
			}
		case Chat:
		case GoldCoin:
		case Inform:
		case Goodbye:
			return
		}
	}
}

func TestImpose(t *testing.T) {
	imp := Message{Hello, nil, "world"}
	noodle := make(chan Message)
	p2p := make(chan Message)
	go pirateListen(noodle) // pirate to be controlled directly by noodly touch
	go pirateListen(p2p)    // pirate that will get messages via the other one
	// test p2p by telling first pirate about the other
	Message{NameDrop, p2p, "Mate"}.GoSend(noodle)
	// test all options including namedrop nil and goodbye
	for i := 0; i < int(numOfImpositions); i++ {
		imp.Imposition = Impositions(i)
		noodle <- imp
	}
	// shut down second pirate, which will have said hello twice
	p2p <- Message{Goodbye, nil, "Pasta la vista"}
}
