// protocol support for go providing a way to send a variety of commands
// and types over a single channel by encoding the type

package gotocol

type Impositions int

// message types to be imposed on the receiver
const (
	// Hello ChanToFSM NameForPirate // initial noodly touch
	Hello Impositions = iota
	// Namedrop ChanToBuddy NameOfBuddy // here's someone to talk to
	NameDrop
	// Chat - "ThisOften:1.0" // chat to buddies every X seconds
	Chat
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
	case Goodbye:
		return "Goodbye"
	}
	return "Unknown"
}

// structure used for all messages, includes a channel of itself
type Message struct {
	Imposition   Impositions  // request type
	ResponseChan chan Message // place to send more messages
	Intention    string       // payload
}

func Send(to chan<- Message, msg Message) {
	to <- msg
}
