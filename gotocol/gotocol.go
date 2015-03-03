// Package gotocol provides protocol support to send a variety of commands
// listener channels and types over a single channel type
package gotocol

import (
	"fmt"
	"time"
)

// Impositions is the promise theory term for requests made to a service
type Impositions int

// Constant definitions for message types to be imposed on the receiver
const (
	// Hello ChanToParent Name Initial noodly touch to set identity
	Hello Impositions = iota
	// NameDrop ChanToBuddy NameOfBuddy Here's someone to talk to
	NameDrop
	// Chat - ThisOften Chat to buddies time interval
	Chat
	// GoldCoin FromChan HowMuch
	GoldCoin
	// Inform loggerChan text message
	Inform
	// GetRequest FromChan key Simulate http inbound request
	GetRequest
	// GetResponse FromChan value Simulate http outbound response
	GetResponse
	// Put - "key value" Save the key and value
	Put
	// Replicate - "key value" Save a replicated copy
	Replicate
	// Forget - NameOfBuddy Forget connection to buddy
	Forget
	// Delete - key Remove key and value
	Delete
	// Goodbye - - // tell FSM and exit
	Goodbye // test assumes this is the last and exits
	numOfImpositions
)

// String handler to make imposition types printable
func (imps Impositions) String() string {
	switch imps {
	case Hello:
		return "Hello"
	case NameDrop:
		return "NameDrop"
	case Chat:
		return "Chat"
	case GoldCoin:
		return "GoldCoin"
	case Inform:
		return "Inform"
	case GetRequest:
		return "GetRequest"
	case GetResponse:
		return "GetResponse"
	case Put:
		return "Put"
	case Replicate:
		return "Replicate"
	case Forget:
		return "Forget"
	case Delete:
		return "Delete"
	case Goodbye:
		return "Goodbye"
	}
	return "Unknown"
}

// Message structure used for all messages, includes a channel of itself
type Message struct {
	Imposition   Impositions  // request type
	ResponseChan chan Message // place to send response messages
	Sent         time.Time    // time at which message was sent
	Intention    string       // payload
}

func (msg Message) String() string {
	return fmt.Sprintf("gotocol: %v %v %v", time.Since(msg.Sent), msg.Imposition, msg.Intention)
}

// Send a synchronous message
func Send(to chan<- Message, msg Message) {
	to <- msg
}

// GoSend asynchronous message send, parks it on a new goroutine until it completes
func (msg Message) GoSend(to chan Message) {
	go func(c chan Message, m Message) { c <- m }(to, msg)
}
