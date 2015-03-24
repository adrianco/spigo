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
	"github.com/adrianco/spigo/monolith"       // business logic monolith
	"github.com/adrianco/spigo/names"          // manage service name hierarchy
	"github.com/adrianco/spigo/pirate"         // random end user network
	"github.com/adrianco/spigo/priamCassandra" // Priam managed Cassandra cluster
	"github.com/adrianco/spigo/staash"         // storage tier as a service http - data access layer
	"github.com/adrianco/spigo/store"          // generic storage service
	"github.com/adrianco/spigo/zuul"           // API proxy microservice router
	"log"
	"time"
)

var (
	// noodles channels mapped by microservice name connects netflixoss to everyone
	noodles    map[string]chan gotocol.Message
	eurekachan map[string]chan gotocol.Message // eureka for each region.zone
	listener   chan gotocol.Message            // netflixoss listener
	root       string                          // root name to run

	// AWS region names
	rnames = [...]string{"us-east-1", "us-west-2", "eu-west-1", "eu-east-1", "ap-south-1", "ap-south-2"}
	// netflixoss always needs three zones
	znames = [...]string{"zoneA", "zoneB", "zoneC"}
)

// Create a tier
func Create(servicename, packagename string, regions, count int, dependencies ...string) string {
	var name string
	if regions == 0 { // for dns that isn't in a region or zone
		log.Printf("Create cross region: " + servicename)
		name = names.Make(archaius.Conf.Arch, "*", "*", servicename, packagename, 0)
		StartPackage(name, dependencies)
	}
	for r := 0; r < regions; r++ {
		if count == 0 { // for AWS services that are cross zone like elb
			log.Printf("Create cross zone: " + servicename)
			name = names.Make(archaius.Conf.Arch, rnames[r], "*", servicename, packagename, 0)
			StartPackage(name, dependencies)
		} else {
			log.Printf("Create service: " + servicename)
			for i := r * count; i < (r+1)*count; i++ {
				name = names.Make(archaius.Conf.Arch, rnames[r], znames[i%3], servicename, packagename, i)
				StartPackage(name, dependencies)
			}
		}
	}
	return name
}

func StartPackage(name string, dependencies []string) {
	if names.Package(name) == "eureka" {
		eurekachan[name] = make(chan gotocol.Message, archaius.Conf.Population)
		go eureka.Start(eurekachan[name], name)
		return
	} else {
		noodles[name] = make(chan gotocol.Message)
	}
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
	case "monolith":
		go monolith.Start(noodles[name])
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
	// there is a eureka service registry in each zone, so in-zone services just get to talk to their local registry
	// elb are cross zone, so need to see all registries in a region
	// denominator are cross region so need to see all registries globally
	for n, ch := range eurekachan {
		if names.Region(name) == "*" {
			// need to know every eureka in all zones and regions
			noodles[name] <- gotocol.Message{gotocol.Inform, ch, time.Now(), n}
		} else {
			if names.Zone(name) == "*" && names.Region(name) == names.Region(n) {
				// need every eureka in my region
				noodles[name] <- gotocol.Message{gotocol.Inform, ch, time.Now(), n}
			} else {
				if names.RegionZone(name) == names.RegionZone(n) {
					// just the eureka in this specific zone
					noodles[name] <- gotocol.Message{gotocol.Inform, ch, time.Now(), n}
				}
			}
		}
	}
	// pass on symbolic dependencies without channels that will be looked up in Eureka later
	for _, dep := range dependencies {
		if dep != "" {
			noodles[name] <- gotocol.Message{gotocol.NameDrop, nil, time.Now(), dep}
		}
	}
}

// Reload the network from a file
func Reload(arch string) {
	listener = make(chan gotocol.Message) // listener for netflixoss
	log.Println("migration reloading from " + arch + ".json")
	g := graphjson.ReadArch(arch)
	archaius.Conf.Population = 0 // just to make sure
	// count how many nodes there are
	for _, element := range g.Graph {
		if element.Node != "" {
			archaius.Conf.Population++
		}
	}
	// create the maps of channels
	noodles = make(map[string]chan gotocol.Message, archaius.Conf.Population)
	eurekachan = make(map[string]chan gotocol.Message, 3*archaius.Conf.Regions)

	// eureka and edda aren't recorded in the json file to simplify the graph
	// Start all the services
	for _, element := range g.Graph {
		if element.Node != "" {
			name := element.Node
			StartPackage(name, nil)
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
	listener = make(chan gotocol.Message) // listener for netflixoss
	if archaius.Conf.Population < 1 {
		log.Fatal("migration: can't create less than 1 microservice")
	} else {
		log.Printf("migration: scaling to %v%%", archaius.Conf.Population)
	}
	// create maps of channels
	noodles = make(map[string]chan gotocol.Message, archaius.Conf.Population)
	eurekachan = make(map[string]chan gotocol.Message, 3*archaius.Conf.Regions)

	// Build the configuration step by step

	// create eureka service registries in each zone
	euname := "eureka"
	eucount := 3
	// start mysql data store layer, which connects to itself
	mysqlcount := 2
	sname := "rds-mysql"
	// start memcached layer, only one per region
	mname := "memcache"
	mcount := 1
	if archaius.Conf.StopStep >= 3 {
		// start evcache layer, one per zone
		mname = "evcache"
		mcount = 3
	}
	// priam managed Cassandra cluster, turtle because it's used to configure other clusters
	priamCassandracount := 12 * archaius.Conf.Population / 100
	cname := "cassTurtle"
	cpkg := "priamCassandra"
	// staash data access layer connects to mysql master and slave, and evcache
	staashcount := 6 * archaius.Conf.Population / 100
	tname := "turtle"
	//  php business logic, we can create a network of simple services from the karyon package
	phpcount := 9 * archaius.Conf.Population / 100
	pname := "php"
	// some node microservice logic, we can create a network of simple services from the karyon package
	nodecount := 9 * archaius.Conf.Population / 100
	nname := "node"
	// zuul api proxies and insert between elb and php
	zuulcount := 9 * archaius.Conf.Population / 100
	zuname := "wwwproxy"
	// AWS elastic load balancer
	elbname := "www-elb"
	// DNS endpoint
	dns := "www"

	// setup name service and cross zone replication links
	Create(euname, "eureka", archaius.Conf.Regions, eucount)
	for n, ch := range eurekachan {
		var n1, n2 string
		switch names.Zone(n) {
		case znames[0]:
			n1 = znames[1]
			n2 = znames[2]
		case znames[1]:
			n1 = znames[0]
			n2 = znames[2]
		case znames[2]:
			n1 = znames[0]
			n2 = znames[1]
		}
		for nn, cch := range eurekachan {
			if names.Region(nn) == names.Region(n) && (names.Zone(nn) == n1 || names.Zone(nn) == n2) {
				log.Println("Eureka cross connect from: " + n + " to " + nn)
				ch <- gotocol.Message{gotocol.NameDrop, cch, time.Now(), nn}
			}
		}
	}

	switch archaius.Conf.StopStep {
	case 1: // basic LAMP with memcache
		Create(sname, "store", archaius.Conf.Regions, mysqlcount, sname)
		Create(mname, "store", archaius.Conf.Regions, mcount)
		Create(pname, "monolith", archaius.Conf.Regions, phpcount, sname, mname)
		Create(elbname, "elb", archaius.Conf.Regions, 0, pname)
	case 2: // LAMP with zuul and memcache
		Create(sname, "store", archaius.Conf.Regions, mysqlcount, sname)
		Create(mname, "store", archaius.Conf.Regions, mcount)
		Create(pname, "monolith", archaius.Conf.Regions, phpcount, sname, mname)
		Create(zuname, "zuul", archaius.Conf.Regions, zuulcount, pname)
		Create(elbname, "elb", archaius.Conf.Regions, 0, zuname)
	case 3: // LAMP with zuul and staash and evcache
		Create(sname, "store", archaius.Conf.Regions, mysqlcount, sname)
		Create(mname, "store", archaius.Conf.Regions, mcount)
		Create(tname, "staash", archaius.Conf.Regions, staashcount, sname, mname)
		Create(pname, "karyon", archaius.Conf.Regions, phpcount, tname)
		Create(zuname, "zuul", archaius.Conf.Regions, zuulcount, pname)
		Create(elbname, "elb", archaius.Conf.Regions, 0, zuname)
	case 4: // added node microservice
		Create(sname, "store", archaius.Conf.Regions, mysqlcount, sname)
		Create(mname, "store", archaius.Conf.Regions, mcount)
		Create(tname, "staash", archaius.Conf.Regions, staashcount, sname, mname, cname)
		Create(pname, "karyon", archaius.Conf.Regions, phpcount, tname)
		Create(nname, "karyon", archaius.Conf.Regions, nodecount, tname)
		Create(zuname, "zuul", archaius.Conf.Regions, zuulcount, pname, nname)
		Create(elbname, "elb", archaius.Conf.Regions, 0, zuname)
	case 5: // added cassandra alongside mysql
		Create(cname, "priamCassandra", archaius.Conf.Regions, priamCassandracount, cname)
		Create(sname, "store", archaius.Conf.Regions, mysqlcount, sname)
		Create(mname, "store", archaius.Conf.Regions, mcount)
		Create(tname, "staash", archaius.Conf.Regions, staashcount, sname, mname, cname)
		Create(pname, "karyon", archaius.Conf.Regions, phpcount, tname)
		Create(nname, "karyon", archaius.Conf.Regions, nodecount, tname)
		Create(zuname, "zuul", archaius.Conf.Regions, zuulcount, pname, nname)
		Create(elbname, "elb", archaius.Conf.Regions, 0, zuname)
	default: // for all higher steps
		fallthrough
	case 6: // removed mysql so that multi-region will work properly
		Create(cname, "priamCassandra", archaius.Conf.Regions, priamCassandracount, cname)
		Create(mname, "store", archaius.Conf.Regions, mcount)
		Create(tname, "staash", archaius.Conf.Regions, staashcount, mname, cname)
		Create(pname, "karyon", archaius.Conf.Regions, phpcount, tname)
		Create(nname, "karyon", archaius.Conf.Regions, nodecount, tname)
		Create(zuname, "zuul", archaius.Conf.Regions, zuulcount, pname, nname)
		Create(elbname, "elb", archaius.Conf.Regions, 0, zuname)
	}

	dnsname := Create(dns, "denominator", 0, 0, elbname)

	// stop here for for single region, then add second region, then join them
	if archaius.Conf.StopStep < 8 {
		run(dnsname)
		return
	}
	// Connect cross region Cassandra0
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
	// shutdown eureka and wait to catch eureka reply
	for _, ch := range eurekachan {
		gotocol.Message{gotocol.Goodbye, listener, time.Now(), "shutdown"}.GoSend(ch)
	}
	for _ = range eurekachan {
		msg = <-listener
	}
	// wait for all the eureka to flush messages and exit
	eureka.Wg.Wait()
	collect.Save()
	log.Println("migration: Exit")
}
