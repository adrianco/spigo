spigo
=====

Simulate Protocol Interactions in Go

Uses a common message type Gotocol which contains a channel of the same type. This allows message listener endpoints to be passed around to dynamically create an arbitrary interconnection network.

Using terminology from Promise Theory each message also has an Imposition code that tells the receiver how to interpret it, and an Intention body string that can be used as a simple string, or to encode a more complex structured type or a Promise.

There is a central controller, the FSM, and a number of independent Pirates who listen to the FSM and to each other.

Initial implementation creates the FSM and ten pirates, the FSM sends a Hello PirateNN message to name them which includes the FSM listener channel for back-chat. FSM then sends a Goodbye message to each, the Pirate then quits and confirms by sending a Goodbye message back to the FSM.

There is a test program that exercises the Namedrop message, this is where the FSM or a Pirate passes on the name of a third party, and each Pirate builds up a buddy list of names and the listener channel for each buddy.

The basic framework is in place, but the interesting behaviors haven't been added yet.
