// Tests for gotocol
package gotocol

import (
	"fmt"
	"testing"
	"time"
)

// all configuration and state is sent via messages
func pirateListen(listener chan Message) {
	var buddy chan Message
	var msg Message
	for {
		msg = <-listener
		fmt.Println(listener, msg)
		// handle all message types
		switch msg.Imposition {
		case Hello:
		case NameDrop:
			// remember the buddy for later if we got one
			if msg.ResponseChan != nil {
				buddy = msg.ResponseChan
			}
		case Chat:
			// send  a Request if we have a buddy
			if buddy != nil {
				go Send(buddy, Message{GetRequest, listener, time.Now(), "Yo ho ho"})
			}
		case GoldCoin:
		case Inform:
		case GetRequest:
			if msg.ResponseChan != nil {
				Message{GetResponse, nil, time.Now(), "Bottle of rum"}.GoSend(msg.ResponseChan)
			}
		case GetResponse:
		case Put:
		case Forget:
		case Delete:
		case Goodbye:
			return
		}
	}
}

func TestImpose(t *testing.T) {
	imp := Message{Hello, nil, time.Now(), "world"}
	noodle := make(chan Message)
	p2p := make(chan Message)
	go pirateListen(noodle) // pirate to be controlled directly by noodly touch
	go pirateListen(p2p)    // pirate that will get messages via the other one
	// test p2p by telling first pirate about the other
	Message{NameDrop, p2p, time.Now(), "Mate"}.GoSend(noodle)
	// test all options including namedrop nil and goodbye
	for i := 0; i < int(numOfImpositions); i++ {
		imp.Imposition = Impositions(i)
		noodle <- imp
	}
	// shut down second pirate, which will have said hello twice
	p2p <- Message{Goodbye, nil, time.Now(), "Pasta la vista"}
}
