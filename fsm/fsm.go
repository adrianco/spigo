// Flexible State Manager (a.k.a. Flying Spaghetti Monster)
// Controls everything, touching with its noodles
package fsm

import (
	"fmt"
	"github.com/adrianco/spigo/gotocol"
	"math/rand"
	"time"
)

var ChatSleep time.Duration

// Touch touches all the noodles that connect to the pirates etc.
func Touch(noodles map[string]chan gotocol.Message) {
	var msg gotocol.Message
	names := make([]string, len(noodles)) // indexable name list
	listener := make(chan gotocol.Message)
	fmt.Println("Hello")
	i := 0
	for name, noodle := range noodles {
		noodle <- gotocol.Message{gotocol.Hello, listener, name}
		names[i] = name
		i++
	}
	fmt.Println("Talk amongst yourselves for", ChatSleep)
	rand.Seed(int64(len(noodles)))
	start := time.Now()
	for _, name := range names {
		ch := noodles[name]
		// for each pirate tell them about two other random pirates
		talkto := names[rand.Intn(len(names))]
		ch <- gotocol.Message{gotocol.NameDrop, noodles[talkto], talkto}
		talkto = names[rand.Intn(len(names))]
		ch <- gotocol.Message{gotocol.NameDrop, noodles[talkto], talkto}
		ch <- gotocol.Message{gotocol.Chat, nil, "2s"}
	}
	d := time.Since(start)
	fmt.Println("Delivered", 3*len(names), "messages in", d)
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
		if msg.Imposition == gotocol.Goodbye {
			delete(noodles, msg.Intention)
			fmt.Printf("Pirate population: %v    \r", len(noodles))
		}
	}
	fmt.Println("\nExit")
}
