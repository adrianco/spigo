// protocol support for go providing a way to send a variety of commands
// and types over a single channel by encoding the type

package gotocol

type Impositions int

// message types to be imposed on the reciever
const (
	Hello Impositions = iota
	NameDrop
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
