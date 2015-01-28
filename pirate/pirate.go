// Package pirate is a participant in the social network, listens to the FSM and to other pirates
// chats at a variable rate by giving gold and namedropping friends
package pirate

import (
	"fmt"
	"github.com/adrianco/spigo/gotocol"
	"log"
	"math/rand"
	"time"
)

// Msglog turns on console logging of messages
var Msglog bool

// Start the pirate, all configuration and state is sent via messages
func Start(listener chan gotocol.Message) {
	dunbar := 10 // starting point for how many buddies to remember
	// remember the channel to talk to named buddies
	buddies := make(map[string]chan gotocol.Message, dunbar)
	// remember who sent GoldCoin and how much, to buy favors
	benefactors := make(map[string]int, dunbar)
	var booty int                   // current GoldCoin balance
	var fsm chan gotocol.Message    // remember how to talk back to creator
	var name string                 // remember my name
	var logger chan gotocol.Message // if set, send updates
	var chatrate time.Duration
	chatTicker := time.NewTicker(time.Hour)
	chatTicker.Stop()
	for {
		select {
		case msg := <-listener:
			if Msglog {
				log.Printf("%v: %v\n", name, msg)
			}
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
				// logger channel is buffered so no need to use GoSend
				logger <- gotocol.Message{gotocol.Hello, nil, name + " " + "pirate"}
			case gotocol.NameDrop:
				// don't remember too many buddies and don't talk to myself
				buddy := msg.Intention // message body is buddy name
				if len(buddies) < dunbar && buddy != name {
					// remember how to talk to this buddy
					buddies[buddy] = msg.ResponseChan // message channel is buddy's listener
					if logger != nil {
						// if it's setup, tell the logger I have a new buddy to talk to
						logger <- gotocol.Message{gotocol.Inform, listener, name + " " + buddy}
					}
				}
			case gotocol.Chat:
				// setup the ticker to run at the specified rate
				d, e := time.ParseDuration(msg.Intention)
				if e == nil && d >= time.Millisecond && d <= time.Hour {
					chatrate = d
					chatTicker = time.NewTicker(chatrate)
					// assume we got paid before we started chatting
					rand.Seed(int64(booty))
				}
			case gotocol.GoldCoin:
				var coin int
				_, e := fmt.Sscanf(msg.Intention, "%d", &coin)
				if e == nil && coin > 0 {
					booty += coin
					for name, ch := range buddies {
						if ch == msg.ResponseChan {
							benefactors[name] += coin
						}
					}
				}
			case gotocol.Goodbye:
				if Msglog {
					log.Printf("%v: Going away with %v gold coins, chatting every %v\n", name, booty, chatrate)
				}
				gotocol.Message{gotocol.Goodbye, nil, name}.GoSend(fsm)
				return
			}
		case <-chatTicker.C:
			if rand.Intn(100) < 50 { // 50% of the time
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
			} else {
				// send a buddy some money
				if booty > 0 {
					donation := rand.Intn(booty)
					luckyNumber := rand.Intn(len(buddies))
					if donation > 0 {
						for _, ch := range buddies {
							if luckyNumber == 0 {
								gotocol.Message{gotocol.GoldCoin, listener, fmt.Sprintf("%d", donation)}.GoSend(ch)
								booty -= donation
								break
							} else {
								luckyNumber--
							}
						}
					}
				}
			}
		}
	}
}
