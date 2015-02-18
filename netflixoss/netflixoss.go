// Package netflixoss implements a simulation of a global large scale microservice architecture
// It creates and controls a collection of aws, netflixoss and netflix application microservices
// or reads in a network from a json file. It also logs the architecture (nodes and links) as it evolves
package netflixoss

import (
	"fmt"
	"github.com/adrianco/spigo/archaius"    // global configuration
	"github.com/adrianco/spigo/collect"     // metrics collector
	"github.com/adrianco/spigo/denominator" // DNS service
	"github.com/adrianco/spigo/elb"         // elastic load balancer
	"github.com/adrianco/spigo/eureka"      // service and attribute registry
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
	go eureka.Start(eurekachan, "netflixoss.eureka")
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
			case "denominator":
				go denominator.Start(noodles[name])
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
	go eureka.Start(eurekachan, "netflixoss.eureka")

	// we need a DNS service to create a global multi-region architecture
	dnsname := "netflixoss.global-api-dns"
	noodles[dnsname] = make(chan gotocol.Message)
	go denominator.Start(noodles[dnsname])
	// setup the dns name and logging, set chat rate after everything else is started
	noodles[dnsname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), dnsname}
	noodles[dnsname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}

	// we need elb as a front end in each region to spread request traffic around each endpoint
	rnames := [...]string{"us-east-1", "us-west-2", "eu-west-1", "eu-east-1", "ap-south-1", "ap-south-2"}
	// remember them in the global config
	for i, s := range rnames {
		archaius.Conf.RegionNames[i] = s
	}
	// netflixoss always needs three zones
	znames := [...]string{"zoneA", "zoneB", "zoneC"}
	// elb for api endpoint
	for r := 0; r < archaius.Conf.Regions; r++ {
		rname := "netflixoss." + rnames[r]
		elbname := fmt.Sprintf("%v-elb", rname)
		noodles[elbname] = make(chan gotocol.Message)
		go elb.Start(noodles[elbname])
		// setup the elb's name and logging, set chat rate after everything else is started
		noodles[elbname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), elbname}
		noodles[elbname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
		// tell denominator how to talk to the elb
		noodles[dnsname] <- gotocol.Message{gotocol.NameDrop, noodles[elbname], time.Now(), elbname}

		// start zuul api proxies next
		zuulcount := 9 * archaius.Conf.Population / 100
		for i := r * zuulcount; i < (r+1)*zuulcount; i++ {
			zuulname := fmt.Sprintf("%v.%v.zuul%v", rname, znames[i%3], i)
			noodles[zuulname] = make(chan gotocol.Message)
			go zuul.Start(noodles[zuulname])
			noodles[zuulname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), zuulname}
			zone := fmt.Sprintf("zone %v.%v", rname, znames[i%3])
			noodles[zuulname] <- gotocol.Message{gotocol.Put, nil, time.Now(), zone}
			noodles[zuulname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
			// hook all the zuul proxies up to the elb
			noodles[elbname] <- gotocol.Message{gotocol.NameDrop, noodles[zuulname], time.Now(), zuulname}
		}

		// start karyon business logic
		karyoncount := 27 * archaius.Conf.Population / 100
		for i := r * karyoncount; i < (r+1)*karyoncount; i++ {
			karyonname := fmt.Sprintf("%v.%v.karyon%v", rname, znames[i%3], i)
			noodles[karyonname] = make(chan gotocol.Message)
			go karyon.Start(noodles[karyonname])
			noodles[karyonname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), karyonname}
			zone := fmt.Sprintf("zone %v.%v", rname, znames[i%3])
			noodles[karyonname] <- gotocol.Message{gotocol.Put, nil, time.Now(), zone}
			noodles[karyonname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
			// connect all the karyon in a zone to all zuul in that zone only
			for j := r*zuulcount + i%3; j < (r+1)*zuulcount; j = j + 3 {
				zuul := fmt.Sprintf("%v.%v.zuul%v", rname, znames[i%3], j)
				noodles[zuul] <- gotocol.Message{gotocol.NameDrop, noodles[karyonname], time.Now(), karyonname}
			}
		}
		// start staash data access layer
		staashcount := 6 * archaius.Conf.Population / 100
		for i := r * staashcount; i < (r+1)*staashcount; i++ {
			staashname := fmt.Sprintf("%v.%v.staash%v", rname, znames[i%3], i)
			noodles[staashname] = make(chan gotocol.Message)
			go staash.Start(noodles[staashname])
			noodles[staashname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), staashname}
			zone := fmt.Sprintf("zone %v.%v", rname, znames[i%3])
			noodles[staashname] <- gotocol.Message{gotocol.Put, nil, time.Now(), zone}
			noodles[staashname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
			// connect all the staash in a zone to all karyon in that zone only
			for j := r*karyoncount + i%3; j < (r+1)*karyoncount; j = j + 3 {
				karyon := fmt.Sprintf("%v.%v.karyon%v", rname, znames[i%3], j)
				noodles[karyon] <- gotocol.Message{gotocol.NameDrop, noodles[staashname], time.Now(), staashname}
			}
		}
		// start priam managed Cassandra cluster
		priamCassandracount := 12 * archaius.Conf.Population / 100
		for i := r * priamCassandracount; i < (r+1)*priamCassandracount; i++ {
			priamCassandraname := fmt.Sprintf("%v.%v.priamCassandra%v", rname, znames[i%3], i)
			noodles[priamCassandraname] = make(chan gotocol.Message)
			go priamCassandra.Start(noodles[priamCassandraname])
			noodles[priamCassandraname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), priamCassandraname}
			zone := fmt.Sprintf("zone %v.%v", rname, znames[i%3])
			noodles[priamCassandraname] <- gotocol.Message{gotocol.Put, nil, time.Now(), zone}
			noodles[priamCassandraname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
			// connect all the priamCassandra in a zone to all staash in that zone only
			for j := r*staashcount + i%3; j < (r+1)*staashcount; j = j + 3 {
				staash := fmt.Sprintf("%v.%v.staash%v", rname, znames[i%3], j)
				noodles[staash] <- gotocol.Message{gotocol.NameDrop, noodles[priamCassandraname], time.Now(), priamCassandraname}
			}
		}
		// make the cross zone priamCassandra connections, assumes staash/astayanax ring aware client routing
		for i := r * priamCassandracount; i < (r+1)*priamCassandracount; i++ {
			priamCassandraZ0 := fmt.Sprintf("%v.%v.priamCassandra%v", rname, znames[i%3], i)
			priamCassandraZ1 := fmt.Sprintf("%v.%v.priamCassandra%v", rname, znames[(i+1)%3], r*priamCassandracount+(i+1)%priamCassandracount)
			noodles[priamCassandraZ0] <- gotocol.Message{gotocol.NameDrop, noodles[priamCassandraZ1], time.Now(), priamCassandraZ1}
			priamCassandraZ2 := fmt.Sprintf("%v.%v.priamCassandra%v", rname, znames[(i+2)%3], r*priamCassandracount+(i+2)%priamCassandracount)
			noodles[priamCassandraZ0] <- gotocol.Message{gotocol.NameDrop, noodles[priamCassandraZ2], time.Now(), priamCassandraZ2}
		}
	}

	// Connect cross region Cassandra
	priamCassandracount := 12 * archaius.Conf.Population / 100
	if archaius.Conf.Regions > 1 {
		for r := 0; r < archaius.Conf.Regions; r++ {
			for i := r * priamCassandracount; i < (r+1)*priamCassandracount; i++ {
				for j := 1; j < archaius.Conf.Regions; j++ {
					pC := fmt.Sprintf("netflixoss.%v.%v.priamCassandra%v", rnames[r], znames[i%3], i)
					pCindex := (i + priamCassandracount) % (archaius.Conf.Regions * priamCassandracount)
					pCremote := fmt.Sprintf("netflixoss.%v.%v.priamCassandra%v", rnames[(r+1)%archaius.Conf.Regions], znames[pCindex%3], pCindex)
					//log.Printf("%v %v\n", pC, pCremote)
					noodles[pC] <- gotocol.Message{gotocol.NameDrop, noodles[pCremote], time.Now(), pCremote}
				}
			}
		}
	}

	// tell denominator to start chatting with microservices every 0.01 secs
	delay := fmt.Sprintf("%dms", 10)
	log.Println("netflixoss: denominator activity rate ", delay)
	noodles[dnsname] <- gotocol.Message{gotocol.Chat, nil, time.Now(), delay}
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
	// listen one extra time to catch eureka reply
	gotocol.Message{gotocol.Goodbye, listener, time.Now(), "shutdown"}.GoSend(eurekachan)
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
	// wait for eureka to flush messages and exit
	eureka.Wg.Wait()
	collect.Save()
	log.Println("netflixoss: Exit")
}
