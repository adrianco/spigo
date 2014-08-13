// Flexible State Manager (a.k.a. Flying Spaghetti Monster)
// Controls everything, touching with its noodles

package fsm

import (
	"fmt"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphml"
	"math/rand"
	"time"
)

var ChatSleep time.Duration

// FSM touches all the noodles that connect to the pirates etc.
func Touch(noodles map[string]chan gotocol.Message) {
	var msg gotocol.Message
	names := make([]string, len(noodles)) // indexable name list
	listener := make(chan gotocol.Message)
	graphml.Setup()
	fmt.Println("Hello")
	i := 0
	for name, noodle := range noodles {
		graphml.WriteNode(name)
		noodle <- gotocol.Message{gotocol.Hello, listener, name}
		names[i] = name
		i = i + 1
	}
	fmt.Println("Talk amongst yourselves for", ChatSleep)
	rand.Seed(int64(len(noodles)))
	start := time.Now()
	for i := 0; i < len(names); i++ {
		// for each pirate tell them about two other random pirates
		noodle := noodles[names[i]] // lookup the channel
		// pick a first random pirate to tell this one about
		talkto := names[rand.Intn(len(names))]
		noodle <- gotocol.Message{gotocol.NameDrop, noodles[talkto], talkto}
		// pick a second random pirate to tell this one about
		talkto = names[rand.Intn(len(names))]
		noodle <- gotocol.Message{gotocol.NameDrop, noodles[talkto], talkto}
		// send this pirate a random amount of GoldCoin up to 100
		gold := fmt.Sprintf("%d", rand.Intn(100))
		noodle <- gotocol.Message{gotocol.GoldCoin, listener, gold}
		// tell this pirate to start chatting with friends every 1-10s
		delay := fmt.Sprintf("%ds", 1+rand.Intn(9))
		noodle <- gotocol.Message{gotocol.Chat, nil, delay}
	}
	d := time.Since(start)
	fmt.Println("Delivered", 4*len(names), "messages in", d)
	if ChatSleep >= time.Millisecond {
		time.Sleep(ChatSleep)
	}
	fmt.Println("Go away")
	for _, noodle := range noodles {
		noodle <- gotocol.Message{gotocol.Goodbye, nil, "beer volcano"}
	}
	for len(noodles) > 0 {
		msg = <-listener
		// fmt.Println(msg)
		switch msg.Imposition {
		case gotocol.Inform:
			graphml.Write(msg.Intention)
		case gotocol.Goodbye:
			delete(noodles, msg.Intention)
			fmt.Printf("Pirate population: %v    \r", len(noodles))
		}
	}
	fmt.Println("\nExit")
	graphml.Close()
}
