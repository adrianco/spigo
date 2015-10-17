// package handlers contains common code used for message handling
package handlers

import (
	"github.com/adrianco/spigo/flow"
	"github.com/adrianco/spigo/gotocol"
	"math/rand"
	"time"
)

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
		(*requestor)[outmsg.Ctx.String()] = msg.Route() // remember where to respond to when this span comes back
		outmsg.GoSend((*microindex)[m])
	}
}

// Responsehandler provides generic response handling
func GetResponse(msg gotocol.Message, name string, listener chan gotocol.Message, requestor *map[string]gotocol.Routetype) {
	ctx := msg.Ctx.String()
	r := (*requestor)[ctx]
	if r.ResponseChan != nil {
		outmsg := gotocol.Message{gotocol.GetResponse, listener, time.Now(), r.Ctx, msg.Intention}
		flow.AnnotateSend(outmsg, name)
		outmsg.GoSend(r.ResponseChan)
		delete(*requestor, ctx)
	}
}
