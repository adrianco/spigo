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
	"github.com/adrianco/spigo/names"          // manage service name hierarchy
	"github.com/adrianco/spigo/pirate"         // random end user network
	"github.com/adrianco/spigo/priamCassandra" // Priam managed Cassandra cluster
	"github.com/adrianco/spigo/staash"         // storage tier as a service http - data access layer
	"github.com/adrianco/spigo/zuul"           // API proxy microservice router
	"log"
	"time"
)

// noodles channels mapped by microservice name connects netflixoss to everyone
var noodles map[string]chan gotocol.Message

var listener chan gotocol.Message   // netflixoss listener
var eurekachan chan gotocol.Message // eureka - eventually for each zone and region

// Reload the network from a file
func Reload(arch string) {
	var root string                                                   // root name to run
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
	// eureka and edda aren't recorded in the json file to simplify the graph
	// TODO need to have a eureka per region
	go eureka.Start(eurekachan, "netflixoss.eureka")
	// Start all the services
	for _, element := range g.Graph {
		if element.Node != "" {
			name := element.Node
			noodles[name] = make(chan gotocol.Message)
			// start the service and tell it it's name
			switch names.Package(name) {
			case "pirate":
				go pirate.Start(noodles[name])
			case "elb":
				go elb.Start(noodles[name])
			case "denominator":
				go denominator.Start(noodles[name])
				root = name
			case "zuul":
				go zuul.Start(noodles[name])
			case "karyon":
				go karyon.Start(noodles[name])
			case "staash":
				go staash.Start(noodles[name])
			case "priamCassandra":
				go priamCassandra.Start(noodles[name])
			default:
				log.Fatal("netflixoss: unknown package: " + names.Package(name))
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
	// run for a while
	if root == "" {
		log.Fatal("No denominator microservice specified")
	}
	run(root)
}

// Start netflixoss and create new microservices
func Start() {
	arch := "netflixoss"
	listener = make(chan gotocol.Message)                             // listener for netflixoss
	eurekachan = make(chan gotocol.Message, archaius.Conf.Population) // listener for netflixoss
	if archaius.Conf.Population < 1 {
		log.Fatal("netflixoss: can't create less than 1 microservice")
	} else {
		log.Printf("netflixoss: scaling to %v%%", archaius.Conf.Population)
	}
	// create map of channels
	noodles = make(map[string]chan gotocol.Message, archaius.Conf.Population)

	// start the service registry first, TODO needs to be one per region
	go eureka.Start(eurekachan, "netflixoss.eureka")

	// we need a DNS service to create a global multi-region architecture
	dnsname := "netflixoss.global-api-dns"
	noodles[dnsname] = make(chan gotocol.Message)
	go denominator.Start(noodles[dnsname])
	// setup the dns name and logging, set chat rate after everything else is started
	noodles[dnsname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), dnsname}
	noodles[dnsname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
	// pause at this point to generate partial config for demonstrations
	if archaius.Conf.StopStep == 1 {
		run(dnsname)
		return
	}

	// we need elb as a front end in each region to spread request traffic around each endpoint
	rnames := [...]string{"us-east-1", "us-west-2", "eu-west-1", "eu-east-1", "ap-south-1", "ap-south-2"}
	// remember them in the global config
	for i, s := range rnames {
		archaius.Conf.RegionNames[i] = s
	}
	// netflixoss always needs three zones
	znames := [...]string{"zoneA", "zoneB", "zoneC"}
	// elb for api endpoint
	elbcnt := 0
	// cross region cassandra cluster names need to scope outside loop
	cname := "cassTurtle"
	cpkg := "priamCassandra"
	for r := 0; r < archaius.Conf.Regions; r++ {
		rname := rnames[r]
		elbname := names.Make(arch, rname, "ABC", "api-elb", "elb", elbcnt)
		elbcnt++
		noodles[elbname] = make(chan gotocol.Message)
		go elb.Start(noodles[elbname])
		// setup the elb's name and logging, set chat rate after everything else is started
		noodles[elbname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), elbname}
		noodles[elbname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
		// tell denominator how to talk to the elb
		noodles[dnsname] <- gotocol.Message{gotocol.NameDrop, noodles[elbname], time.Now(), elbname}
		if archaius.Conf.StopStep == 2 {
			run(dnsname)
			return
		}

		// start zuul api proxies next
		zuulcount := 9 * archaius.Conf.Population / 100
		zuname := "apiproxy"
		zupkg := "zuul"
		for i := r * zuulcount; i < (r+1)*zuulcount; i++ {
			zuulname := names.Make(arch, rname, znames[i%3], zuname, zupkg, i)
			noodles[zuulname] = make(chan gotocol.Message)
			go zuul.Start(noodles[zuulname])
			noodles[zuulname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), zuulname}
			noodles[zuulname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
			// hook all the zuul proxies up to the elb in this region
			noodles[elbname] <- gotocol.Message{gotocol.NameDrop, noodles[zuulname], time.Now(), zuulname}
		}
		if archaius.Conf.StopStep == 3 {
			run(dnsname)
			return
		}

		// start api business logic, we can create a network of simple services from the karyon package
		apicount := 27 * archaius.Conf.Population / 100
		aname := "api"
		apkg := "karyon"
		for i := r * apicount; i < (r+1)*apicount; i++ {
			apiname := names.Make(arch, rname, znames[i%3], aname, apkg, i)
			noodles[apiname] = make(chan gotocol.Message)
			go karyon.Start(noodles[apiname])
			noodles[apiname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), apiname}
			noodles[apiname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
			// connect all the api in a zone to all zuul in that zone only
			for j := r*zuulcount + i%3; j < (r+1)*zuulcount; j = j + 3 {
				zuul := names.Make(arch, rname, znames[i%3], zuname, zupkg, j)
				noodles[zuul] <- gotocol.Message{gotocol.NameDrop, noodles[apiname], time.Now(), apiname}
			}
		}
		if archaius.Conf.StopStep == 4 {
			run(dnsname)
			return
		}

		// start staash data access layer
		staashcount := 6 * archaius.Conf.Population / 100
		sname := "turtle"
		spkg := "staash"
		for i := r * staashcount; i < (r+1)*staashcount; i++ {
			staashname := names.Make(arch, rname, znames[i%3], sname, spkg, i)
			noodles[staashname] = make(chan gotocol.Message)
			go staash.Start(noodles[staashname])
			noodles[staashname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), staashname}
			noodles[staashname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
			// connect all the staash in a zone to all api in that zone only
			for j := r*apicount + i%3; j < (r+1)*apicount; j = j + 3 {
				api := names.Make(arch, rname, znames[i%3], aname, apkg, j)
				noodles[api] <- gotocol.Message{gotocol.NameDrop, noodles[staashname], time.Now(), staashname}
			}
		}
		if archaius.Conf.StopStep == 5 {
			run(dnsname)
			return
		}

		// start first priam managed Cassandra cluster, turtle because it's used to configure other clusters
		priamCassandracount := 12 * archaius.Conf.Population / 100
		for i := r * priamCassandracount; i < (r+1)*priamCassandracount; i++ {
			priamCassandraname := names.Make(arch, rname, znames[i%3], cname, cpkg, i)
			noodles[priamCassandraname] = make(chan gotocol.Message)
			go priamCassandra.Start(noodles[priamCassandraname])
			noodles[priamCassandraname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), priamCassandraname}
			noodles[priamCassandraname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
			// connect all the priamCassandra in a zone to all staash in that zone only
			for j := r*staashcount + i%3; j < (r+1)*staashcount; j = j + 3 {
				staash := names.Make(arch, rname, znames[i%3], sname, spkg, j)
				noodles[staash] <- gotocol.Message{gotocol.NameDrop, noodles[priamCassandraname], time.Now(), priamCassandraname}
			}
		}
		if archaius.Conf.StopStep == 6 {
			run(dnsname)
			return
		}

		// make the cross zone priamCassandra connections, assumes staash/astayanax ring aware client routing
		for i := r * priamCassandracount; i < (r+1)*priamCassandracount; i++ {
			priamCassandraZ0 := names.Make(arch, rname, znames[i%3], cname, cpkg, i)
			priamCassandraZ1 := names.Make(arch, rname, znames[(i+1)%3], cname, cpkg, r*priamCassandracount+(i+1)%priamCassandracount)
			noodles[priamCassandraZ0] <- gotocol.Message{gotocol.NameDrop, noodles[priamCassandraZ1], time.Now(), priamCassandraZ1}
			priamCassandraZ2 := names.Make(arch, rname, znames[(i+2)%3], cname, cpkg, r*priamCassandracount+(i+2)%priamCassandracount)
			noodles[priamCassandraZ0] <- gotocol.Message{gotocol.NameDrop, noodles[priamCassandraZ2], time.Now(), priamCassandraZ2}
		}
	}

	// stop here for 7 for single region, then add second region for step 8, then join them for 9
	if archaius.Conf.StopStep == 7 || archaius.Conf.StopStep == 8 {
		run(dnsname)
		return
	}
	// Connect cross region Cassandra
	priamCassandracount := 12 * archaius.Conf.Population / 100
	if archaius.Conf.Regions > 1 {
		// for each region
		for r := 0; r < archaius.Conf.Regions; r++ {
			// for each priamCassandrian in that region
			for i := r * priamCassandracount; i < (r+1)*priamCassandracount; i++ {
				pC := names.Make(arch, rnames[r], znames[i%3], cname, cpkg, i)
				// for each of the other regions connect to one node
				for j := 1; j < archaius.Conf.Regions; j++ {
					pCindex := (i + j*priamCassandracount) % (archaius.Conf.Regions * priamCassandracount)
					pCremote := names.Make(arch, rnames[(r+1)%archaius.Conf.Regions], znames[pCindex%3], cname, cpkg, pCindex)
					noodles[pC] <- gotocol.Message{gotocol.NameDrop, noodles[pCremote], time.Now(), pCremote}
				}
			}
		}
	}
	run(dnsname)
}

// Run netflixoss for a while then shut down
func run(root string) {
	var msg gotocol.Message
	// tell denominator to start chatting with microservices every 0.01 secs
	delay := fmt.Sprintf("%dms", 10)
	log.Println("netflixoss: denominator activity rate ", delay)
	noodles[root] <- gotocol.Message{gotocol.Chat, nil, time.Now(), delay}

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
