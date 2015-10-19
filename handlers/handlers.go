// package handlers contains common code used for message handling
package handlers

import (
	"github.com/adrianco/spigo/flow"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/names"
	"log"
	"math/rand"
	"time"
)

// InformHandler default handler for Inform message
func Inform(msg gotocol.Message, name string, listener chan gotocol.Message) chan gotocol.Message {
	if name == "" {
		log.Fatal(name + "Inform message received before Hello message")
	}
	// service registry channel is buffered so don't use GoSend to tell Eureka we exist
	msg.ResponseChan <- gotocol.Message{gotocol.Put, listener, time.Now(), gotocol.NilContext, name}
	return msg.ResponseChan
}

// NameDrop updates local buddy list
func NameDrop(dependencies *map[string]time.Time, microservices *map[string]chan gotocol.Message, msg gotocol.Message, name string, listener chan gotocol.Message, eureka map[string]chan gotocol.Message, crosszone ...bool) {
	if msg.ResponseChan == nil { // dependency by service name, needs to be looked up in eureka
		(*dependencies)[msg.Intention] = msg.Sent // remember it for later
		for _, ch := range eureka {
			//log.Println(name + " looking up " + msg.Intention)
			gotocol.Send(ch, gotocol.Message{gotocol.GetRequest, listener, time.Now(), gotocol.NilContext, msg.Intention})
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
					gotocol.Send(ch, gotocol.Message{gotocol.Inform, listener, time.Now(), gotocol.NilContext, name + " " + microservice})
					return
				}
			}
		}
	}
}

// Forget removes a buddy from the buddy list
func Forget(dependencies *map[string]time.Time, microservices *map[string]chan gotocol.Message, msg gotocol.Message) {
	microservice := msg.Intention              // message body is buddy name to forget
	if (*microservices)[microservice] != nil { // an existing buddy to forget
		// forget how to talk to this buddy
		(*dependencies)[names.Service(microservice)] = msg.Sent // remember when we were told to forget this service
		delete(*microservices, microservice)
	}
}

func Put(msg gotocol.Message, name string, listener chan gotocol.Message, requestor *map[string]gotocol.Routetype, microservices *map[string]chan gotocol.Message, microindex *map[int]chan gotocol.Message) {
	if len(*microservices) > 0 {
		if len(*microindex) != len(*microservices) {
			// rebuild index
			i := 0
			for _, ch := range *microservices {
				(*microindex)[i] = ch
				i++
			}
		}
		m := rand.Intn(len(*microservices))
		// pass on request to a random service - client send
		outmsg := gotocol.Message{gotocol.Put, listener, time.Now(), msg.Ctx.NewParent(), msg.Intention}
		flow.AnnotateSend(outmsg, name)
		outmsg.GoSend((*microindex)[m])
	}
}

func GetRequest(msg gotocol.Message, name string, listener chan gotocol.Message, requestor *map[string]gotocol.Routetype, microservices *map[string]chan gotocol.Message, microindex *map[int]chan gotocol.Message) {
	if len(*microservices) > 0 {
		if len(*microindex) != len(*microservices) {
			// rebuild index
			i := 0
			for _, ch := range *microservices {
				(*microindex)[i] = ch
				i++
			}
		}
		m := rand.Intn(len(*microservices))
		// pass on request to a random service - client send
		outmsg := gotocol.Message{gotocol.GetRequest, listener, time.Now(), msg.Ctx.NewParent(), msg.Intention}
		flow.AnnotateSend(outmsg, name)
		(*requestor)[outmsg.Ctx.Route()] = msg.Route() // remember where to respond to when this span comes back
		outmsg.GoSend((*microindex)[m])
	}
}

// Responsehandler provides generic response handling
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
