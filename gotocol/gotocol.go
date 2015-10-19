// Package gotocol provides protocol support to send a variety of commands
// listener channels and types over a single channel type
package gotocol

import (
	"fmt"
	"sync/atomic"
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

// trace type needs to be exported for flow package map. For production scale use this should be 64bit, for spigo it seems ok with 32.
type TraceContextType uint32 // needs to match type conversions in func increment below

// context for capturing dapper/zipkin style traces
type Context struct {
	Trace, Parent, Span TraceContextType
}

// string formatter for context
func (ctx Context) String() string {
	return fmt.Sprintf("t%vp%vs%v", ctx.Trace, ctx.Parent, ctx.Span)
}

// string formatter for routing part of context
func (ctx Context) Route() string {
	return fmt.Sprintf("t%vp%v", ctx.Trace, ctx.Parent)
}

// fast hack for generating unique-enough contexts
var spanner TraceContextType

// return uniquely incremented TraceContextType
func increment(tc *TraceContextType) TraceContextType {
	return TraceContextType(atomic.AddUint32((*uint32)(tc), 1))
}

// Start a new trace using atomic increment
func NewTrace() Context {
	var ctx Context
	//ctx.Trace = rand.Uint32()
	// NilContext is t0p0s0, so first real Trace,Parent,Span is t1p0s1
	ctx.Span = increment(&spanner)
	ctx.Trace = ctx.Span // trace = span with zero parent for first span in a trace
	return ctx
}

// Updating to get a new span for an existing request and parent
func (ctx Context) AddSpan() Context {
	ctx.Span = increment(&spanner)
	return ctx
}

// setup the parent by promoting incoming span id, and get a new spanid
func (ctx Context) NewParent() Context {
	ctx.Parent = ctx.Span
	return ctx.AddSpan()
}

// make an empty context, can't figure out how to make this a const
var NilContext Context

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

// Routing information from a message
type Routetype struct {
	Ctx          Context
	ResponseChan chan Message
}

// extract routing information from a message
func (msg Message) Route() Routetype {
	var r Routetype
	r.Ctx = msg.Ctx
	r.ResponseChan = msg.ResponseChan
	return r
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
