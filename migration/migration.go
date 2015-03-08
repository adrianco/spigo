// Package migration implements a simulation of migration to a global large scale microservice architecture
// It creates and controls a collection of aws, lamp, netflixoss and netflix application microservices
// or reads in a network from a json file. It also logs the architecture (nodes and links) as it evolves
package migration

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
	"github.com/adrianco/spigo/store"          // generic storage service
	"github.com/adrianco/spigo/zuul"           // API proxy microservice router
	"log"
	"time"
)

// noodles channels mapped by microservice name connects netflixoss to everyone
var noodles map[string]chan gotocol.Message

var listener chan gotocol.Message   // netflixoss listener
var eurekachan chan gotocol.Message // eureka - eventually for each zone and region
var root string                     // root name to run

// AWS region names
var rnames = [...]string{"us-east-1", "us-west-2", "eu-west-1", "eu-east-1", "ap-south-1", "ap-south-2"}

// netflixoss always needs three zones
var znames = [...]string{"zoneA", "zoneB", "zoneC"}

// Create a tier
func Create(servicename, packagename string, regions, count int, dependencies []string) {
	for r := 0; r < regions; r++ {
		for i := r * count; i < (r+1)*count; i++ {
			name := names.Make(archaius.Conf.Arch, rnames[r], znames[i%3], servicename, packagename, i)
			StartPackage(name)
		}
	}
}

func StartPackage(name string) {
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
	case "store":
		go store.Start(noodles[name])
	default:
		log.Fatal("migration: unknown package: " + names.Package(name))
	}
	noodles[name] <- gotocol.Message{gotocol.Hello, listener, time.Now(), name}
	noodles[name] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
}

// Reload the network from a file
func Reload(arch string) {
	listener = make(chan gotocol.Message)                             // listener for netflixoss
	eurekachan = make(chan gotocol.Message, archaius.Conf.Population) // listener for eureka
	log.Println("migration reloading from " + arch + ".json")
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
			StartPackage(name)
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
		log.Fatal("No denominator root microservice specified")
	}
	run(root)
}

// Start lamp and netflixoss
func Start() {
	arch := archaius.Conf.Arch
	listener = make(chan gotocol.Message)                             // listener for netflixoss
	eurekachan = make(chan gotocol.Message, archaius.Conf.Population) // listener for netflixoss
	if archaius.Conf.Population < 1 {
		log.Fatal("migration: can't create less than 1 microservice")
	} else {
		log.Printf("migration: scaling to %v%%", archaius.Conf.Population)
	}
	// create map of channels
	noodles = make(map[string]chan gotocol.Message, archaius.Conf.Population)

	// start the service registry first, TODO needs to be one per region
	go eureka.Start(eurekachan, "migration.eureka")

	// we need a DNS service to create a global multi-region architecture
	dnsname := names.Make(arch, "*", "*", "www-dns", "denominator", 0)
	noodles[dnsname] = make(chan gotocol.Message)
	go denominator.Start(noodles[dnsname])
	// setup the dns name and logging, set chat rate after everything else is started
	noodles[dnsname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), dnsname}
	noodles[dnsname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}

	// Build the configuration step by step

	// we need elb as a front end in each region to spread request traffic around each endpoint
	// remember regions in the global config
	for i, s := range rnames {
		archaius.Conf.RegionNames[i] = s
	}
	// elb for api endpoint
	elbcnt := 0
	// cross region cassandra cluster names need to scope outside loop
	cname := "cassTurtle"
	cpkg := "priamCassandra"
	for r := 0; r < archaius.Conf.Regions; r++ {
		rname := rnames[r]
		elbname := names.Make(arch, rname, "ABC", "www-elb", "elb", elbcnt)
		elbcnt++
		noodles[elbname] = make(chan gotocol.Message)
		go elb.Start(noodles[elbname])
		// setup the elb's name and logging, set chat rate after everything else is started
		noodles[elbname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), elbname}
		noodles[elbname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
		// tell denominator how to talk to the elb
		noodles[dnsname] <- gotocol.Message{gotocol.NameDrop, noodles[elbname], time.Now(), elbname}

		// start lamp stack

		// start mysql data store layer
		mysqlcount := 2
		sname := "rds-mysql"
		spkg := "store"
		for i := r * mysqlcount; i < (r+1)*mysqlcount; i++ {
			mysqlname := names.Make(arch, rname, znames[i%3], sname, spkg, i)
			noodles[mysqlname] = make(chan gotocol.Message)
			go store.Start(noodles[mysqlname])
			noodles[mysqlname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), mysqlname}
			noodles[mysqlname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
		}
		// connect master mysql in a zone to slave mysql in second zone
		master := names.Make(arch, rname, znames[0], sname, spkg, (r * 2))
		slave := names.Make(arch, rname, znames[1], sname, spkg, (r*2)+1)
		noodles[master] <- gotocol.Message{gotocol.NameDrop, noodles[slave], time.Now(), slave}

		// start memcached layer, one per region
		mname := "memcache"
		mpkg := "store"
		memname := names.Make(arch, rname, znames[0], mname, mpkg, r)
		if archaius.Conf.StopStep < 3 {
			noodles[memname] = make(chan gotocol.Message)
			go store.Start(noodles[memname])
			noodles[memname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), memname}
			noodles[memname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
		}

		// start php business logic, we can create a network of simple services from the karyon package
		phpcount := 9 * archaius.Conf.Population / 100
		pname := "php"
		ppkg := "karyon" // karyon randomly calls its dependencies which isn't really right for master/slave/cache
		for i := r * phpcount; i < (r+1)*phpcount; i++ {
			phpname := names.Make(arch, rname, znames[i%3], pname, ppkg, i)
			noodles[phpname] = make(chan gotocol.Message)
			go karyon.Start(noodles[phpname])
			noodles[phpname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), phpname}
			noodles[phpname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
			if archaius.Conf.StopStep == 1 {
				noodles[elbname] <- gotocol.Message{gotocol.NameDrop, noodles[phpname], time.Now(), phpname}
			}
			if archaius.Conf.StopStep < 3 {
				// connect all the php to mysql and memcached in one zone only
				noodles[phpname] <- gotocol.Message{gotocol.NameDrop, noodles[master], time.Now(), master}
				noodles[phpname] <- gotocol.Message{gotocol.NameDrop, noodles[slave], time.Now(), slave}
				noodles[phpname] <- gotocol.Message{gotocol.NameDrop, noodles[memname], time.Now(), memname}
			}
		}

		if archaius.Conf.StopStep == 1 {
			run(dnsname)
			return
		}

		// start zuul api proxies and insert between elb and php
		zuulcount := 9 * archaius.Conf.Population / 100
		zuname := "wwwproxy"
		zupkg := "zuul"
		for i := r * zuulcount; i < (r+1)*zuulcount; i++ {
			zuulname := names.Make(arch, rname, znames[i%3], zuname, zupkg, i)
			noodles[zuulname] = make(chan gotocol.Message)
			go zuul.Start(noodles[zuulname])
			noodles[zuulname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), zuulname}
			noodles[zuulname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
			// connect all the zuul in a zone to all php in that zone only, and unhook php from elb
			for j := r*phpcount + i%3; j < (r+1)*phpcount; j = j + 3 {
				p := names.Make(arch, rname, znames[i%3], pname, ppkg, j)
				noodles[zuulname] <- gotocol.Message{gotocol.NameDrop, noodles[p], time.Now(), p}
				//noodles[elbname] <- gotocol.Message{gotocol.Forget, nil, time.Now(), p}
			}
			// hook all the zuul proxies up to the elb in this region
			noodles[elbname] <- gotocol.Message{gotocol.NameDrop, noodles[zuulname], time.Now(), zuulname}
		}

		if archaius.Conf.StopStep == 2 {
			run(dnsname)
			return
		}

		// start evcache layer, one per zone
		evcachecount := 3
		mname = "evcache"
		mpkg = "store"
		for i := r * evcachecount; i < (r+1)*evcachecount; i++ {
			evname := names.Make(arch, rname, znames[i%3], mname, mpkg, i)
			noodles[evname] = make(chan gotocol.Message)
			go store.Start(noodles[evname])
			noodles[evname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), evname}
			noodles[evname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
		}

		// start staash data access layer and connect to mysql master and slave, and evcache
		staashcount := 6 * archaius.Conf.Population / 100
		tname := "turtle"
		tpkg := "staash"
		for i := r * staashcount; i < (r+1)*staashcount; i++ {
			staashname := names.Make(arch, rname, znames[i%3], tname, tpkg, i)
			noodles[staashname] = make(chan gotocol.Message)
			go staash.Start(noodles[staashname])
			noodles[staashname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), staashname}
			noodles[staashname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
			// connect to mysql
			noodles[staashname] <- gotocol.Message{gotocol.NameDrop, noodles[master], time.Now(), master}
			noodles[staashname] <- gotocol.Message{gotocol.NameDrop, noodles[slave], time.Now(), slave}
			// connect all staash to all evcache
			for j := 0; j < evcachecount; j++ {
				evname := names.Make(arch, rname, znames[j%3], mname, mpkg, r*evcachecount+j)
				noodles[staashname] <- gotocol.Message{gotocol.NameDrop, noodles[evname], time.Now(), evname}
			}
		}

		// connect php to staash
		for i := r * phpcount; i < (r+1)*phpcount; i++ {
			phpname := names.Make(arch, rname, znames[i%3], pname, ppkg, i)
			// connect all the php in a zone to all staash in that zone only
			for j := r*staashcount + i%3; j < (r+1)*staashcount; j = j + 3 {
				s := names.Make(arch, rname, znames[i%3], tname, tpkg, j)
				noodles[phpname] <- gotocol.Message{gotocol.NameDrop, noodles[s], time.Now(), s}
			}
			// disconnect php from direct access to mysql
			//noodles[phpname] <- gotocol.Message{gotocol.Forget, nil, time.Now(), master}
			//noodles[phpname] <- gotocol.Message{gotocol.Forget, nil, time.Now(), slave}
		}

		if archaius.Conf.StopStep == 3 {
			run(dnsname)
			return
		}

		// start more microservice logic, we can create a network of simple services from the karyon package
		nodecount := 9 * archaius.Conf.Population / 100
		nname := "node"
		npkg := "karyon"
		for i := r * nodecount; i < (r+1)*nodecount; i++ {
			nodename := names.Make(arch, rname, znames[i%3], nname, npkg, i)
			noodles[nodename] = make(chan gotocol.Message)
			go karyon.Start(noodles[nodename])
			noodles[nodename] <- gotocol.Message{gotocol.Hello, listener, time.Now(), nodename}
			noodles[nodename] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
			for j := r*staashcount + i%3; j < (r+1)*staashcount; j = j + 3 {
				s := names.Make(arch, rname, znames[i%3], tname, tpkg, j)
				noodles[nodename] <- gotocol.Message{gotocol.NameDrop, noodles[s], time.Now(), s}
			}
		}
		for i := r * zuulcount; i < (r+1)*zuulcount; i++ {
			zuulname := names.Make(arch, rname, znames[i%3], zuname, zupkg, i)
			// connect all the zuul in a zone to all node in that zone only
			for j := r*nodecount + i%3; j < (r+1)*nodecount; j = j + 3 {
				n := names.Make(arch, rname, znames[i%3], nname, npkg, j)
				noodles[zuulname] <- gotocol.Message{gotocol.NameDrop, noodles[n], time.Now(), n}
			}
		}

		if archaius.Conf.StopStep == 4 {
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
		}

		// make the cross zone priamCassandra connections, assumes staash/astayanax ring aware client routing
		for i := r * priamCassandracount; i < (r+1)*priamCassandracount; i++ {
			priamCassandraZ0 := names.Make(arch, rname, znames[i%3], cname, cpkg, i)
			priamCassandraZ1 := names.Make(arch, rname, znames[(i+1)%3], cname, cpkg, r*priamCassandracount+(i+1)%priamCassandracount)
			noodles[priamCassandraZ0] <- gotocol.Message{gotocol.NameDrop, noodles[priamCassandraZ1], time.Now(), priamCassandraZ1}
			priamCassandraZ2 := names.Make(arch, rname, znames[(i+2)%3], cname, cpkg, r*priamCassandracount+(i+2)%priamCassandracount)
			noodles[priamCassandraZ0] <- gotocol.Message{gotocol.NameDrop, noodles[priamCassandraZ2], time.Now(), priamCassandraZ2}
		}

		if archaius.Conf.StopStep == 5 {
			run(dnsname)
			return
		}

		// connect staash data access layer to Cassandra
		for i := r * staashcount; i < (r+1)*staashcount; i++ {
			staashname := names.Make(arch, rname, znames[i%3], tname, tpkg, i)
			// connect all the staash in a zone to all priamCassandra in that zone only
			for j := r*priamCassandracount + i%3; j < (r+1)*priamCassandracount; j = j + 3 {
				pc := names.Make(arch, rname, znames[i%3], cname, cpkg, j)
				noodles[staashname] <- gotocol.Message{gotocol.NameDrop, noodles[pc], time.Now(), pc}
			}
		}
		if archaius.Conf.StopStep == 6 {
			run(dnsname)
			return
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

// Run migration for a while then shut down
func run(rootservice string) {
	var msg gotocol.Message
	// tell denominator to start chatting with microservices every 0.01 secs
	delay := fmt.Sprintf("%dms", 10)
	log.Println("migration: denominator activity rate ", delay)
	noodles[rootservice] <- gotocol.Message{gotocol.Chat, nil, time.Now(), delay}

	// wait until the delay has finished
	if archaius.Conf.RunDuration >= time.Millisecond {
		time.Sleep(archaius.Conf.RunDuration)
	}
	log.Println("migration: Shutdown")
	for _, noodle := range noodles {
		gotocol.Message{gotocol.Goodbye, nil, time.Now(), "shutdown"}.GoSend(noodle)
	}
	// listen one extra time to catch eureka reply
	gotocol.Message{gotocol.Goodbye, listener, time.Now(), "shutdown"}.GoSend(eurekachan)
	for len(noodles) > 0 {
		msg = <-listener
		if archaius.Conf.Msglog {
			log.Printf("migration: %v\n", msg)
		}
		switch msg.Imposition {
		case gotocol.Goodbye:
			delete(noodles, msg.Intention)
			if archaius.Conf.Msglog {
				log.Printf("migration: %v shutdown, population: %v    \n", msg.Intention, len(noodles))
			}
		}
	}
	// wait for eureka to flush messages and exit
	eureka.Wg.Wait()
	collect.Save()
	log.Println("migration: Exit")
}
