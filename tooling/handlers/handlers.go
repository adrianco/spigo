// Package handlers contains common code used for message handling
package handlers

import (
	"github.com/adrianco/spigo/tooling/archaius"
	"github.com/adrianco/spigo/tooling/flow"
	"github.com/adrianco/spigo/tooling/gotocol"
	"github.com/adrianco/spigo/tooling/names"
	"github.com/adrianco/spigo/tooling/ribbon"
	"log"
	"time"
)

// DebugContext turns on debug context logging for eureka and edda messages
func DebugContext(ctx gotocol.Context) gotocol.Context {
	if archaius.Conf.Msglog && archaius.Conf.Collect {
		// combination of -m and -c command line creates msglog and records flow as zipkin or (with -n) neo4j
		if ctx == gotocol.NilContext {
			// start of a trace
			return gotocol.NewTrace()
		}
		// next step of an existing trace
		return ctx.NewParent()
	}
	return gotocol.NilContext
}

// Inform default handler for Inform message
func Inform(msg gotocol.Message, name string, listener chan gotocol.Message) chan gotocol.Message {
	if name == "" {
		log.Fatal(name + "Inform message received before Hello message")
	}
	// service registry channel is buffered so don't use GoSend to tell Eureka we exist
	msg.ResponseChan <- gotocol.Message{gotocol.Put, listener, time.Now(), DebugContext(msg.Ctx), name}
	return msg.ResponseChan
}

// NameDrop updates local buddy list
func NameDrop(dependencies *map[string]time.Time, router *ribbon.Router, msg gotocol.Message, name string, listener chan gotocol.Message, eureka map[string]chan gotocol.Message, crosszone ...bool) {
	if msg.ResponseChan == nil { // dependency by service name, needs to be looked up in eureka
		(*dependencies)[msg.Intention] = msg.Sent // remember it for later
		for _, ch := range eureka {
			//log.Println(name + " looking up " + msg.Intention)
			gotocol.Send(ch, gotocol.Message{gotocol.GetRequest, listener, time.Now(), DebugContext(msg.Ctx), msg.Intention})
		}
	} else { // update dependency with full name and listener channel
		microservice := msg.Intention // message body is buddy name
		if len(crosszone) > 0 || names.Zone(name) == names.Zone(microservice) {
			if microservice != name && router.Named(microservice) == nil { // don't talk to myself or record duplicates
				// remember how to talk to this buddy
				router.Add(microservice, msg.ResponseChan, msg.Sent) // message channel is buddy's listener
				(*dependencies)[names.Service(microservice)] = msg.Sent
				for _, ch := range eureka {
					// tell just one of the service registries I have a new buddy to talk to so it doesn't get logged more than once
					gotocol.Send(ch, gotocol.Message{gotocol.Inform, listener, time.Now(), DebugContext(msg.Ctx), name + " " + microservice})
					return
				}
			}
		}
	}
}

// Forget removes a buddy from the buddy list
func Forget(dependencies *map[string]time.Time, router *ribbon.Router, msg gotocol.Message) {
	microservice := msg.Intention          // message body is buddy name to forget
	if router.Named(microservice) != nil { // an existing buddy to forget
		// forget how to talk to this buddy
		(*dependencies)[names.Service(microservice)] = msg.Sent // remember when we were told to forget this service
		router.Remove(microservice)
	}
}

// Put sends a Put message to a service
func Put(msg gotocol.Message, name string, listener chan gotocol.Message, requestor *map[string]gotocol.Routetype, router *ribbon.Router) {
	// pass on request to a random service - client send
	c := router.Random()
	if c == nil {
		return
	}
	outmsg := gotocol.Message{gotocol.Put, listener, time.Now(), msg.Ctx.NewParent(), msg.Intention}
	flow.AnnotateSend(outmsg, name)
	outmsg.GoSend(c)
}

// GetRequest sends a GetRequest message to a service
func GetRequest(msg gotocol.Message, name string, listener chan gotocol.Message, requestor *map[string]gotocol.Routetype, router *ribbon.Router) {
	// pass on request to a random service - client send
	c := router.Random()
	if c == nil {
		return
	}
	outmsg := gotocol.Message{gotocol.GetRequest, listener, time.Now(), msg.Ctx.NewParent(), msg.Intention}
	flow.AnnotateSend(outmsg, name)
	(*requestor)[outmsg.Ctx.Route()] = msg.Route() // remember where to respond to when this span comes back
	outmsg.GoSend(c)
}

// GetResponse provides generic response handling
func GetResponse(msg gotocol.Message, name string, listener chan gotocol.Message, requestor *map[string]gotocol.Routetype) {
	ctr := msg.Ctx.Route()
	r := (*requestor)[ctr]
	if r.ResponseChan != nil {
		outmsg := gotocol.Message{gotocol.GetResponse, listener, time.Now(), r.Ctx, msg.Intention}
		flow.AnnotateSend(outmsg, name)
		outmsg.GoSend(r.ResponseChan)
		delete(*requestor, ctr)
	}
}
