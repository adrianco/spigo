// Package gotocol provides protocol support to send a variety of commands
// listener channels and types over a single channel type
package gotocol

import (
	"fmt"
	"github.com/adrianco/spigo/names"
	"log"
	//"math/rand"
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
	// Forget - FromBuddy ToBuddy Forget link between two buddies
	Forget
	// Delete - key Remove key and value
	Delete
	// Goodbye - name // tell FSM and exit
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

// context for capturing dapper/zipkin style traces
type Context struct {
	request, span uint32
}

// string formatter for context
func (ctx Context) String() string {
	return fmt.Sprintf("%v:%v", ctx.request, ctx.span)
}

// fast hack for generating unique-enough contexts
var spanner, requester uint32

// new request happens less often so use random for request, and increment span
func NewRequest() Context {
	var ctx Context
	//ctx.request = rand.Uint32()
	requester++
	ctx.request = requester // NilContext is 0:0, so first real Context is 1:0
	ctx.span = spanner
	spanner++
	return ctx
}

// updating to get a new span for an existing request
func (ctx Context) NewSpan() Context {
	ctx.span = spanner
	spanner++
	return ctx
}

// make an empty context
func NilContext() Context {
	return Context{0, 0}
}

// Message structure used for all messages, includes a channel of itself
type Message struct {
	Imposition   Impositions  // request type
	ResponseChan chan Message // place to send response messages
	Sent         time.Time    // time at which message was sent
	Ctx          Context      // message context
	Intention    string       // payload
}

func (msg Message) String() string {
	return fmt.Sprintf("gotocol: %v %v %v %v", time.Since(msg.Sent), msg.Ctx, msg.Imposition, msg.Intention)
}

// Send a synchronous message
func Send(to chan<- Message, msg Message) {
	if to != nil {
		to <- msg
	}
}

// GoSend asynchronous message send, parks it on a new goroutine until it completes
func (msg Message) GoSend(to chan Message) {
	go func(c chan Message, m Message) {
		if c != nil {
			c <- m
		}
	}(to, msg)
}

// InformHandler default handler for Inform message
func InformHandler(msg Message, name string, listener chan Message) chan Message {
	if name == "" {
		log.Fatal(name + "Inform message received before Hello message")
	}
	// service registry channel is buffered so don't use GoSend to tell Eureka we exist
	msg.ResponseChan <- Message{Put, listener, time.Now(), NilContext(), name}
	return msg.ResponseChan
}

func NameDropHandler(dependencies *map[string]time.Time, microservices *map[string]chan Message, msg Message, name string, listener chan Message, eureka map[string]chan Message, crosszone ...bool) {
	if msg.ResponseChan == nil { // dependency by service name, needs to be looked up in eureka
		(*dependencies)[msg.Intention] = msg.Sent // remember it for later
		for _, ch := range eureka {
			//log.Println(name + " looking up " + msg.Intention)
			Send(ch, Message{GetRequest, listener, time.Now(), NilContext(), msg.Intention})
		}
	} else { // update dependency with full name and listener channel
		microservice := msg.Intention // message body is buddy name
		if len(crosszone) > 0 || names.Zone(name) == names.Zone(microservice) {
			if microservice != name && (*microservices)[microservice] == nil { // don't talk to myself or record duplicates
				// remember how to talk to this buddy
				(*microservices)[microservice] = msg.ResponseChan // message channel is buddy's listener
				(*dependencies)[names.Service(microservice)] = msg.Sent
				for _, ch := range eureka {
					// tell one of the service registries I have a new buddy to talk to so it doesn't get logged more than once
					Send(ch, Message{Inform, listener, time.Now(), NilContext(), name + " " + microservice})
					return
				}
			}
		}
	}
}

// ForgetHandler removes a buddy from the buddy list
func ForgetHandler(dependencies *map[string]time.Time, microservices *map[string]chan Message, msg Message) {
	microservice := msg.Intention              // message body is buddy name to forget
	if (*microservices)[microservice] != nil { // an existing buddy to forget
		// forget how to talk to this buddy
		(*dependencies)[names.Service(microservice)] = msg.Sent // remember when we were told to forget this service
		delete(*microservices, microservice)
	}
}
