// Flexible State Manager (a.k.a. Flying Spaghetti Monster)
// Controls everything, touching with its noodles

package fsm

import (
	"github.com/adrianco/spigo/gotocol"
	"fmt"
)

// FSM touches all the noodles that connect to the pirates etc.
func Touch(noodles map[string]chan gotocol.Message) {
	var msg gotocol.Message
	listener := make(chan gotocol.Message)
	fmt.Println("Hello")
	for name, noodle := range noodles {
		noodle <- gotocol.Message{gotocol.Hello, listener, name}
	}
	fmt.Println("Go away")
	for _, noodle := range noodles {
                noodle <- gotocol.Message{gotocol.Goodbye, nil, "beer volcano"}
        }
	for len(noodles) > 0 {
		msg = <-listener
		fmt.Println(msg)
		if msg.Imposition == gotocol.Goodbye {
			delete(noodles, msg.Intention)
			fmt.Println(len(noodles), " pirates left")
		}
	}	

}

