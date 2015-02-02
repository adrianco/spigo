// Package gotocol provides protocol support to send a variety of commands
// listener channels and types over a single channel type
package gotocol

// Impositions is the promise theory term for requests made to a service
type Impositions int

// Constant definitions for message types to be imposed on the receiver
const (
	// Hello ChanToParent NameForPirate // initial noodly touch
	Hello Impositions = iota
	// NameDrop ChanToBuddy NameOfBuddy // here's someone to talk to
	NameDrop
	// Chat - ThisOften // chat to buddies time interval
	Chat
	// GoldCoin FromChan HowMuch
	GoldCoin
	// Inform loggerChan text message
	Inform
	// GetRequest FromChan key // simulate http inbound request
	GetRequest
	// GetResponse FromChan value // simulate http outbound response
	GetResponse
	// Put - "key value" // save the key and value
	Put
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
	case Goodbye:
		return "Goodbye"
	}
	return "Unknown"
}

// Message structure used for all messages, includes a channel of itself
type Message struct {
	Imposition   Impositions  // request type
	ResponseChan chan Message // place to send response messages
	Intention    string       // payload
}

// Send a synchronous message
func Send(to chan<- Message, msg Message) {
	to <- msg
}

// GoSend asynchronous message send, parks it on a new goroutine until it completes
func (msg Message) GoSend(to chan Message) {
	go func(c chan Message, m Message) { c <- m }(to, msg)
}
