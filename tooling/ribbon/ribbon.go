//Ribbon is named after the NetflixOSS routing and load balancing project, functions for routing traffic
package ribbon

import (
	"github.com/adrianco/spigo/tooling/gotocol"
	"github.com/adrianco/spigo/tooling/names"
	"math/rand"
	"time"
)

type Router struct {
	routes  map[string]chan gotocol.Message
	updated map[string]time.Time // dependent services and time last updated
}

// Make a new Router with map initialized
func MakeRouter() *Router {
	var r *Router
	r = new(Router)
	r.routes = make(map[string]chan gotocol.Message)
	r.updated = make(map[string]time.Time)
	return r
}

// Size of routing table
func (r *Router) Len() int {
	return len(r.routes)
}

// Add an entry to the routing table
func (r *Router) Add(name string, c chan gotocol.Message, t time.Time) {
	var nilt time.Time
	r.routes[name] = c
	if t != nilt {
		r.updated[name] = t
	}
}

// Remove an entry from the routing table
func (r *Router) Remove(name string) {
	delete(r.routes, name)
	delete(r.updated, name)
}

// Return a random channel from the routing table
func (r *Router) Random() chan gotocol.Message {
	lr := len(r.routes)
	if lr == 0 {
		return nil
	}
	n := rand.Intn(lr)
	for _, c := range r.routes {
		if n == 0 {
			return c
		} else {
			n--
		}
	}
	return nil // the table was empty
}

// Return all routes that match a package
func (r *Router) All(p string) *Router {
	packroutes := MakeRouter()
	var t time.Time
	for n, c := range r.routes {
		if names.Package(n) == p {
			packroutes.Add(n, c, t)
		}
	}
	return packroutes
}

// Pick a random matching package and return that channel from the routing table
func (r *Router) Pick(p string) chan gotocol.Message {
	return r.All(p).Random()
}

// Pick a specific entry and return that channel from the routing table
func (r *Router) Named(n string) chan gotocol.Message {
	return r.routes[n]
}

// Find the name corresponding to a channel
func (r *Router) NameChan(ch chan gotocol.Message) string {
	for n, c := range r.routes {
		if ch == c {
			return n
		}
	}
	return ""
}

// Return all the names
func (r *Router) Names() (ns []string) {
	ns = make([]string, 0, 10)
	for n, _ := range r.routes {
		ns = append(ns, n)
	}
	return ns
}

// Return just the names in the routing table as a string
func (r Router) String() (s string) {
	for _, n := range r.Names() {
		s += (n + " ")
	}
	return s
}
