// test the flow package
package flow

import (
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/gotocol"
	"testing"
	"time"
)

func TestFlow(t *testing.T) {
	archaius.Conf.Collect = true
	archaius.Conf.Arch = "test"
	s1 := gotocol.NewTrace()
	m1 := gotocol.Message{gotocol.GetRequest, nil, time.Now(), s1, "customer1"}
	Annotate(m1, "ss", "requestor", m1.Sent) // BUG can't Annotate(m1...) twice as it overwrites the record
	// pretend there is a subscriber service that got a message from a requestor
	a1 := Annotate(m1, "sr", "subscriber", time.Now())
	s2 := m1.Ctx.NewParent()
	// pretend that the names and addresses services were both sent messages by subscriber
	m2 := gotocol.Message{gotocol.GetRequest, nil, AnnotateSend(a1, s2), s2, m1.Intention}
	s3 := s2.AddSpan()
	m3 := gotocol.Message{gotocol.GetRequest, nil, AnnotateSend(a1, s3), s3, m1.Intention}
	// pretend that names returned to subscriber
	a2 := Annotate(m2, "sr", "names", time.Now())
	s4 := m2.Ctx.NewParent()
	m4 := gotocol.Message{gotocol.GetResponse, nil, AnnotateSend(a2, s4), s4, "name:Fred Flintstone"}
	// pretend that addresses returned to subscriber
	a3 := Annotate(m3, "sr", "addresses", time.Now())
	s5 := m3.Ctx.NewParent()
	m5 := gotocol.Message{gotocol.GetResponse, nil, AnnotateSend(a3, s5), s5, "address:Bedrock"}
	// pretend that subscriber got both messages and joined them together
	Annotate(m4, "sr", "subscriber", time.Now()) // first return doesn't do anything else
	a5 := Annotate(m5, "sr", "subscriber", time.Now())
	s6 := m5.Ctx.NewParent()
	m6 := gotocol.Message{gotocol.GetResponse, nil, AnnotateSend(a5, s6), s6, "name:Fred Flintstone, address:Bedrock"}
	// pretend that requestor got the message
	Annotate(m6, "sr", "requestor", time.Now())

	fmt.Println("All flows")
	fmt.Println(flowmap)
	fmt.Println("\nWalk all remaining flows")
	PrintWalk(flowmap)
	fmt.Println("\nWalk all remaining flows in order to file")
	Walk(flowmap, 0)
	fmt.Println("\nEnd trace 1")
	End(m1.Ctx)
	Shutdown()
}
