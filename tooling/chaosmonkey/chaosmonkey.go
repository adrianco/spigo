// package chaosmonkey deletes nodes
package chaosmonkey

import (
	"github.com/adrianco/spigo/tooling/gotocol"
	"github.com/adrianco/spigo/tooling/names"
	"log"
	"time"
)

// Delete a single node from the given service
func Delete(noodles *map[string]chan gotocol.Message, service string) {
	if service != "" {
		for node, ch := range *noodles {
			if names.Service(node) == service {
				gotocol.Message{gotocol.Goodbye, nil, time.Now(), gotocol.NewTrace(), "chaosmonkey"}.GoSend(ch)
				log.Println("chaosmonkey delete: " + node)
				return
			}
		}
	}
}
