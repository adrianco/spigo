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
				go Send(buddy, Message{GetRequest, listener, time.Now(), NewTrace(), "Yo ho ho"})
			}
		case GoldCoin:
		case Inform:
		case GetRequest:
			if msg.ResponseChan != nil {
				Message{GetResponse, nil, time.Now(), msg.Ctx.NewSpan(), "Bottle of rum"}.GoSend(msg.ResponseChan)
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
	var ctx, ctx4 Context
	ctx2 := NewTrace()
	ctx3 := ctx2.NewSpan()
	ctx4 = NilContext
	fmt.Println("Context: ", ctx, ctx2, ctx3, ctx4, NewTrace())
	imp := Message{Hello, nil, time.Now(), ctx2, "world"}
	noodle := make(chan Message)
	p2p := make(chan Message)
	go pirateListen(noodle) // pirate to be controlled directly by noodly touch
	go pirateListen(p2p)    // pirate that will get messages via the other one
	// test p2p by telling first pirate about the other
	Message{NameDrop, p2p, time.Now(), NilContext, "Mate"}.GoSend(noodle)
	// test all options including namedrop nil and goodbye
	for i := 0; i < int(numOfImpositions); i++ {
		imp.Imposition = Impositions(i)
		imp.Ctx = imp.Ctx.NewSpan()
		noodle <- imp
	}
	// shut down second pirate, which will have said hello twice
	p2p <- Message{Goodbye, nil, time.Now(), NilContext, "Pasta la vista"}
	fmt.Println("len(p2p): ", len(p2p))
	close(p2p)
	fmt.Println("closed len(p2p): ", len(p2p))

}
