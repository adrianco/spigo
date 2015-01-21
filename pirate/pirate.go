// participant in the network, listens to the FSM and to other pirates
// independently decides whether to make or break promises and behave

package pirate

import (
	"fmt"
	"github.com/adrianco/spigo/gotocol"
	"time"
)

// all configuration and state is sent via messages
func Listen(listener chan gotocol.Message) {
	dunbar := 10 // starting point for how many buddies to remember
	// remember the channel to talk to named buddies
	buddies := make(map[string]chan gotocol.Message, dunbar)
	// remember who sent GoldCoin and how much, to buy favors
	benefactors := make(map[string]int, dunbar)
	var booty int                   // current GoldCoin balance
	var fsm chan gotocol.Message    // remember how to talk back to creator
	var name string                 // remember my name
	var logger chan gotocol.Message // if set, send updates
	var msg gotocol.Message
	chatTicker := time.NewTicker(time.Hour)
	chatTicker.Stop()
	for {
		select {
		case msg = <-listener:
			// fmt.Printf("pirate: %v\n", msg)
			switch msg.Imposition {
			case gotocol.Hello:
				if name == "" {
					// if I don't have a name yet remember what I've been named
					fsm = msg.ResponseChan // remember how to talk to my namer
					name = msg.Intention   // message body is my name
				}
			case gotocol.Inform:
				// remember where to send updates
				logger = msg.ResponseChan
			case gotocol.NameDrop:
				// don't remember too many buddies and don't talk to myself
				buddy := msg.Intention // message body is buddy name
				if len(buddies) < dunbar && buddy != name {
					// remember how to talk to this buddy
					buddies[buddy] = msg.ResponseChan // message channel is buddy's listener
					if logger != nil {
						// if it's setup, tell the logger I have a new buddy to talk to
						gotocol.Message{gotocol.Inform, listener, name + " " + buddy}.GoSend(logger)
					}
				}
			case gotocol.Chat:
				// setup the ticker to run at the specified rate
				d, e := time.ParseDuration(msg.Intention)
				if e == nil && d >= time.Millisecond && d <= time.Hour {
					chatTicker = time.NewTicker(d)
				}
			case gotocol.GoldCoin:
				coin, e := fmt.Scanf("%d", msg.Intention)
				if e == nil && coin > 0 {
					booty += coin
					for name, ch := range buddies {
						if ch == msg.ResponseChan {
							benefactors[name] += coin
						}
					}
				}
			case gotocol.Goodbye:
				gotocol.Message{gotocol.Goodbye, nil, name}.GoSend(fsm)
				return
			}
		case _ = <-chatTicker.C:
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
					gotocol.Message{gotocol.NameDrop, firstBuddyChan, firstBuddyName}.GoSend(lastBuddyChan)
				}
			}
		}
	}
}
