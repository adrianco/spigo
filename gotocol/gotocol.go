// protocol support for go providing a way to send a variety of commands
// and types over a single channel by encoding the type

package gotocol

type Impositions int

// message types to be imposed on the receiver
const (
	// Hello ChanToParent NameForPirate // initial noodly touch
	Hello Impositions = iota
	// Namedrop ChanToBuddy NameOfBuddy // here's someone to talk to
	NameDrop
	// Chat - ThisOften // chat to buddies time interval
	Chat
	// GoldCoin FromChan HowMuch
	GoldCoin
	// Inform loggerChan text message
	Inform
	// GetRequest FromChan body // simulate http inbound request
	GetRequest
	// GetResponse FromChan body // simulate http outbound response
	GetResponse
	// Goodbye - - // tell FSM and exit
	Goodbye // test assumes this is the last and exits
	numOfImpositions
)

// make imposition types printable
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
	case Goodbye:
		return "Goodbye"
	}
	return "Unknown"
}

// structure used for all messages, includes a channel of itself
type Message struct {
	Imposition   Impositions  // request type
	ResponseChan chan Message // place to send response messages
	Intention    string       // payload
}

// Send synchronous message
func Send(to chan<- Message, msg Message) {
	to <- msg
}

// GoSend asynchronous message send, parks it on a new goroutine until it completes
func (msg Message) GoSend(to chan Message) {
	go func(c chan Message, m Message) { c <- m }(to, msg)
}
