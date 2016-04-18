// test the flow package
package flow

import (
	"fmt"
	"github.com/adrianco/spigo/tooling/archaius"
	"github.com/adrianco/spigo/tooling/gotocol"
	"testing"
	"time"
)

func TestFlow(t *testing.T) {
	archaius.Conf.Collect = true
	archaius.Conf.Arch = "test"
	s1 := gotocol.NewTrace()
	m1 := gotocol.Message{gotocol.GetRequest, nil, time.Now(), s1, "customer1"}
	AnnotateSend(m1, "requestor")
	// pretend there is a subscriber service that got a message from a requestor
	AnnotateReceive(m1, "subscriber", time.Now())
	s2 := m1.Ctx.NewParent()
	// pretend that the names and addresses services were both sent messages by subscriber
	m2 := gotocol.Message{gotocol.GetRequest, nil, time.Now(), s2, m1.Intention}
	AnnotateSend(m2, "subscriber")
	s3 := s2.AddSpan()
	m3 := gotocol.Message{gotocol.GetRequest, nil, time.Now(), s3, m1.Intention}
	AnnotateSend(m3, "subscriber")
	// pretend that names got the message and returned to subscriber
	AnnotateReceive(m2, "names", time.Now())
	m4 := gotocol.Message{gotocol.GetResponse, nil, time.Now(), m2.Ctx, "name:Fred Flintstone"}
	AnnotateSend(m4, "names")
	// pretend that addresses got the message and returned to subscriber
	AnnotateReceive(m3, "addresses", time.Now())
	m5 := gotocol.Message{gotocol.GetResponse, nil, time.Now(), m3.Ctx, "address:Bedrock"}
	AnnotateSend(m5, "addresses")
	// pretend that subscriber got both messages and joined them together
	AnnotateReceive(m4, "subscriber", time.Now())
	AnnotateReceive(m5, "subscriber", time.Now())
	m6 := gotocol.Message{gotocol.GetResponse, nil, time.Now(), s1, "name:Fred Flintstone, address:Bedrock"}
	AnnotateSend(m6, "subscriber")
	// pretend that requestor got the message
	AnnotateReceive(m6, "requestor", time.Now())

	fmt.Println("All flows")
	for _, f := range flowmap {
		for _, a := range f {
			fmt.Println(*a)
		}
	}
	fmt.Println("\nWrite all remaining flows in order to file")
	Shutdown()
}
