// Flexible State Manager (a.k.a. Flying Spaghetti Monster)
// Controls everything, touching with its noodles

package fsm

import (
	"github.com/adrianco/spigo/gotocol"
	"math/rand"
	"time"
	"fmt"
)

var ChatSleep time.Duration

// FSM touches all the noodles that connect to the pirates etc.
func Touch(noodles map[string]chan gotocol.Message) {
	var msg gotocol.Message
	names := make([]string, len(noodles)) // indexable name list
	listener := make(chan gotocol.Message)
	fmt.Println("Hello")
	i := 0
	for name, noodle := range noodles {
		noodle <- gotocol.Message{gotocol.Hello, listener, name}
		names[i] = name
		i = i + 1
	}
	fmt.Println("Talk amongst yourselves")
	rand.Seed(int64(len(noodles)))
	for i := 0; i < len(names); i++ {
		// for each pirate tell them about two other random pirates
		talkto := names[rand.Intn(len(names))]
		noodles[names[i]] <- gotocol.Message{gotocol.NameDrop, noodles[talkto], talkto} 
		talkto = names[rand.Intn(len(names))]
		noodles[names[i]] <- gotocol.Message{gotocol.NameDrop, noodles[talkto], talkto} 
		noodles[names[i]] <- gotocol.Message{gotocol.Chat, nil, "2s"}
	}
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
	fmt.Println("\nExit");
}

