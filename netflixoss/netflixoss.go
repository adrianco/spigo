// Package netflixoss implements a simulation of a global large scale microservice architecture
// It creates and controls a collection of aws, netflixoss and netflix application microservices
// or reads in a network from a json file. It also logs the architecture (nodes and links) as it evolves
package netflixoss

import (
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/edda"   // configuration logger
	"github.com/adrianco/spigo/elb"    // elastic load balancer
	"github.com/adrianco/spigo/eureka" // service and attribute registry
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/karyon"         // business logic microservice
	"github.com/adrianco/spigo/pirate"         // random end user network
	"github.com/adrianco/spigo/priamCassandra" // Priam managed Cassandra cluster
	"github.com/adrianco/spigo/staash"         // storage tier as a service http - data access layer
	"github.com/adrianco/spigo/zuul"           // API proxy microservice router
	"log"
	"math/rand"
	"time"
)

// noodles channels mapped by microservice name connects netflixoss to everyone
var noodles map[string]chan gotocol.Message
var names []string
var listener, eurekachan chan gotocol.Message

// Reload the network from a file
func Reload(arch string) {
	listener = make(chan gotocol.Message)                             // listener for netflixoss
	eurekachan = make(chan gotocol.Message, archaius.Conf.Population) // listener for eureka
	log.Println("netflixoss reloading from " + arch + ".json")
	g := graphjson.ReadArch(arch)
	archaius.Conf.Population = 0 // just to make sure
	// count how many nodes there are
	for _, element := range g.Graph {
		if element.Node != "" {
			archaius.Conf.Population++
		}
	}
	// create the map of channels
	noodles = make(map[string]chan gotocol.Message, archaius.Conf.Population)
	// Start all the services
	go eureka.Start(eurekachan)
	for _, element := range g.Graph {
		if element.Node != "" && element.Service != "" {
			name := element.Node
			noodles[name] = make(chan gotocol.Message)
			// start the service and tell it it's name
			switch element.Service {
			case "pirate":
				go pirate.Start(noodles[name])
			case "elb":
				go elb.Start(noodles[name])
			case "zuul":
				go zuul.Start(noodles[name])
			case "karyon":
				go karyon.Start(noodles[name])
			case "staash":
				go staash.Start(noodles[name])
			case "priamCassandra":
				go priamCassandra.Start(noodles[name])
			default:
				log.Fatal("netflixoss: unknown service: " + element.Service)
			}
			noodles[name] <- gotocol.Message{gotocol.Hello, listener, time.Now(), name}
			// tell the service to report itself and new edges to the logger
			noodles[name] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
		}
	}
	// Make all the connections
	for _, element := range g.Graph {
		if element.Edge != "" && element.Source != "" && element.Target != "" {
			noodles[element.Source] <- gotocol.Message{gotocol.NameDrop, noodles[element.Target], time.Now(), element.Target}
			log.Println("Link " + element.Source + " > " + element.Target)
		}
	}
	// start the simulation chatting
	for name, noodle := range noodles {
		if name == "elb" {
			// tell each elb to start calling microservices every 0.1 to 1 secs
			delay := fmt.Sprintf("%dms", 100+rand.Intn(900))
			noodle <- gotocol.Message{gotocol.Chat, nil, time.Now(), delay}
		}
	}
	shutdown()
}

// Start netflixoss and create new microservices
func Start() {
	listener = make(chan gotocol.Message)                             // listener for netflixoss
	eurekachan = make(chan gotocol.Message, archaius.Conf.Population) // listener for netflixoss
	if archaius.Conf.Population < 1 {
		log.Fatal("netflixoss: can't create less than 1 microservice")
	} else {
		log.Printf("netflixoss: scaling to %v%%", archaius.Conf.Population)
	}
	// create map of channels and a name index to select randoml nodes from
	noodles = make(map[string]chan gotocol.Message, archaius.Conf.Population)
	names = make([]string, archaius.Conf.Population) // approximate size for indexable name list
	// start the service registry first
	go eureka.Start(eurekachan)
	// we need an elb as a front end to spread request traffic around each endpoint
	// elb for api endpoint
	elbname := "elb-api"
	noodles[elbname] = make(chan gotocol.Message)
	go elb.Start(noodles[elbname])
	// setup the elb's name and logging, set chat rate after everything else is started
	noodles[elbname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), elbname}
	noodles[elbname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
	// connect elb to it's initial dependencies
	// start zuul api proxies next
	zuulcount := 9 * archaius.Conf.Population / 100
	for i := 0; i < zuulcount; i++ {
		zuulname := fmt.Sprintf("zuul%v", i)
		noodles[zuulname] = make(chan gotocol.Message)
		go zuul.Start(noodles[zuulname])
		noodles[zuulname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), zuulname}
		zone := fmt.Sprintf("zone zone%v", i%3)
		noodles[zuulname] <- gotocol.Message{gotocol.Put, nil, time.Now(), zone}
		noodles[zuulname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
		// hook all the zuul proxies up to the elb
		noodles[elbname] <- gotocol.Message{gotocol.NameDrop, noodles[zuulname], time.Now(), zuulname}
	}
	// start karyon business logic
	karyoncount := 27 * archaius.Conf.Population / 100
	for i := 0; i < karyoncount; i++ {
		karyonname := fmt.Sprintf("karyon%v", i)
		noodles[karyonname] = make(chan gotocol.Message)
		go karyon.Start(noodles[karyonname])
		noodles[karyonname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), karyonname}
		zone := fmt.Sprintf("zone zone%v", i%3)
		noodles[karyonname] <- gotocol.Message{gotocol.Put, nil, time.Now(), zone}
		noodles[karyonname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
		// connect all the karyon in a zone to all zuul in that zone only
		for j := i % 3; j < zuulcount; j = j + 3 {
			zuul := fmt.Sprintf("zuul%v", j)
			noodles[zuul] <- gotocol.Message{gotocol.NameDrop, noodles[karyonname], time.Now(), karyonname}
		}
	}
	// start staash data access layer
	staashcount := 6 * archaius.Conf.Population / 100
	for i := 0; i < staashcount; i++ {
		staashname := fmt.Sprintf("staash%v", i)
		noodles[staashname] = make(chan gotocol.Message)
		go staash.Start(noodles[staashname])
		noodles[staashname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), staashname}
		zone := fmt.Sprintf("zone zone%v", i%3)
		noodles[staashname] <- gotocol.Message{gotocol.Put, nil, time.Now(), zone}
		noodles[staashname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
		// connect all the staash in a zone to all karyon in that zone only
		for j := i % 3; j < karyoncount; j = j + 3 {
			karyon := fmt.Sprintf("karyon%v", j)
			noodles[karyon] <- gotocol.Message{gotocol.NameDrop, noodles[staashname], time.Now(), staashname}
		}
	}
	// start priam managed Cassandra cluster
	priamCassandracount := 12 * archaius.Conf.Population / 100
	for i := 0; i < priamCassandracount; i++ {
		priamCassandraname := fmt.Sprintf("priamCassandra%v", i)
		noodles[priamCassandraname] = make(chan gotocol.Message)
		go priamCassandra.Start(noodles[priamCassandraname])
		noodles[priamCassandraname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), priamCassandraname}
		zone := fmt.Sprintf("zone zone%v", i%3)
		noodles[priamCassandraname] <- gotocol.Message{gotocol.Put, nil, time.Now(), zone}
		noodles[priamCassandraname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
		// connect all the priamCassandra in a zone to all staash in that zone only
		for j := i % 3; j < staashcount; j = j + 3 {
			staash := fmt.Sprintf("staash%v", j)
			noodles[staash] <- gotocol.Message{gotocol.NameDrop, noodles[priamCassandraname], time.Now(), priamCassandraname}
		}
	}
	// make the cross zone priamCassandra connections, assumes staash/astayanax ring aware client routing
	for i := 0; i < priamCassandracount; i++ {
		priamCassandraZ0 := fmt.Sprintf("priamCassandra%v", i)
		priamCassandraZ1 := fmt.Sprintf("priamCassandra%v", (i+1)%priamCassandracount)
		noodles[priamCassandraZ0] <- gotocol.Message{gotocol.NameDrop, noodles[priamCassandraZ1], time.Now(), priamCassandraZ1}
		priamCassandraZ2 := fmt.Sprintf("priamCassandra%v", (i+2)%priamCassandracount)
		noodles[priamCassandraZ0] <- gotocol.Message{gotocol.NameDrop, noodles[priamCassandraZ2], time.Now(), priamCassandraZ2}
	}
	// tell this elb to start chatting with microservices every 0.1 secs
	delay := fmt.Sprintf("%dms", 100)
	log.Println("netflixoss: elb activity rate ", delay)
	noodles[elbname] <- gotocol.Message{gotocol.Chat, nil, time.Now(), delay}
	shutdown()
}

// Shutdown netflixoss and elb
func shutdown() {
	var msg gotocol.Message
	// wait until the delay has finished
	if archaius.Conf.RunDuration >= time.Millisecond {
		time.Sleep(archaius.Conf.RunDuration)
	}
	log.Println("netflixoss: Shutdown")
	for _, noodle := range noodles {
		gotocol.Message{gotocol.Goodbye, nil, time.Now(), "shutdown"}.GoSend(noodle)
	}
	for len(noodles) > 0 {
		msg = <-listener
		if archaius.Conf.Msglog {
			log.Printf("netflixoss: %v\n", msg)
		}
		switch msg.Imposition {
		case gotocol.Goodbye:
			delete(noodles, msg.Intention)
			if archaius.Conf.Msglog {
				log.Printf("netflixoss: netflixoss %v shutdown, population: %v    \n", msg.Intention, len(noodles))
			}
		}
	}
	if edda.Logchan != nil {
		close(edda.Logchan)
	}
	log.Println("netflixoss: Exit")
}
