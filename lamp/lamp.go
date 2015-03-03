// Package lamp implements a simulation of a typical LAMP stack
// It creates and controls a collection of aws and LAMP services
// or reads in a network from a json file. It also logs the architecture (nodes and links) as it evolves
package lamp

import (
	"fmt"
	"github.com/adrianco/spigo/archaius"    // global configuration
	"github.com/adrianco/spigo/collect"     // metrics collector
	"github.com/adrianco/spigo/denominator" // DNS service
	"github.com/adrianco/spigo/elb"         // elastic load balancer
	"github.com/adrianco/spigo/eureka"      // service and attribute registry
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/karyon" // business logic microservice
	"github.com/adrianco/spigo/names"  // manage service name hierarchy
	"github.com/adrianco/spigo/store"  // generic storage tier
	"log"
	"time"
)

// noodles channels mapped by microservice name connects lamp to everyone
var noodles map[string]chan gotocol.Message

var listener chan gotocol.Message   // lamp listener
var eurekachan chan gotocol.Message // eureka - eventually for each zone and region

// Reload the network from a file
func Reload(arch string) {
	var root string                                                   // root name to run
	listener = make(chan gotocol.Message)                             // listener for lamp
	eurekachan = make(chan gotocol.Message, archaius.Conf.Population) // listener for eureka
	log.Println("lamp reloading from " + arch + ".json")
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
	go eureka.Start(eurekachan, "lamp.eureka")
	// Start all the services
	for _, element := range g.Graph {
		if element.Node != "" {
			name := element.Node
			noodles[name] = make(chan gotocol.Message)
			// start the service and tell it it's name
			switch names.Package(name) {
			case "elb":
				go elb.Start(noodles[name])
			case "denominator":
				go denominator.Start(noodles[name])
				root = name
			case "karyon":
				go karyon.Start(noodles[name])
			case "store":
				go store.Start(noodles[name])
			default:
				log.Fatal("lamp: unknown package: " + names.Package(name))
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

// Start lamp and create new microservices
func Start() {
	arch := "lamp"
	listener = make(chan gotocol.Message)                             // listener for lamp
	eurekachan = make(chan gotocol.Message, archaius.Conf.Population) // listener for lamp
	if archaius.Conf.Population < 1 {
		log.Fatal("lamp: can't create less than 1 microservice")
	} else {
		log.Printf("lamp: scaling to %v%%", archaius.Conf.Population)
	}
	// create map of channels
	noodles = make(map[string]chan gotocol.Message, archaius.Conf.Population)

	// start the service registry first, TODO needs to be one per region
	go eureka.Start(eurekachan, "lamp.eureka")

	// we need a DNS service to create a global multi-region architecture
	dnsname := names.Make(arch, "*", "*", "www-dns", "denominator", 0)
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

	// Build the configuration bottom up

	// we need elb as a front end in each region to spread request traffic around each endpoint
	rnames := [...]string{"us-east-1", "us-west-2", "eu-west-1", "eu-east-1", "ap-south-1", "ap-south-2"}
	// remember them in the global config
	for i, s := range rnames {
		archaius.Conf.RegionNames[i] = s
	}
	// lamp usually needs two zones
	znames := [...]string{"zoneA", "zoneB"}
	// elb for api endpoint
	elbcnt := 0
	for r := 0; r < archaius.Conf.Regions; r++ {
		rname := rnames[r]
		elbname := names.Make(arch, rname, "AB", "www-elb", "elb", elbcnt)
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

		// start mysql data store layer
		mysqlcount := 2
		sname := "rds-mysql"
		spkg := "store"
		for i := r * mysqlcount; i < (r+1)*mysqlcount; i++ {
			mysqlname := names.Make(arch, rname, znames[i%2], sname, spkg, i)
			noodles[mysqlname] = make(chan gotocol.Message)
			go store.Start(noodles[mysqlname])
			noodles[mysqlname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), mysqlname}
			noodles[mysqlname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
		}
		// connect master mysql in a zone to slave mysql in second zone
		master := names.Make(arch, rname, znames[0], sname, spkg, (r * 2))
		slave := names.Make(arch, rname, znames[1], sname, spkg, (r*2)+1)
		noodles[master] <- gotocol.Message{gotocol.NameDrop, noodles[slave], time.Now(), slave}

		if archaius.Conf.StopStep == 3 {
			run(dnsname)
			return
		}

		// start memcached layer, one per region
		mname := "memcache"
		mpkg := "store"
		memname := names.Make(arch, rname, znames[0], mname, mpkg, r)
		noodles[memname] = make(chan gotocol.Message)
		go store.Start(noodles[memname])
		noodles[memname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), memname}
		noodles[memname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}

		if archaius.Conf.StopStep == 4 {
			run(dnsname)
			return
		}

		// start php business logic, we can create a network of simple services from the karyon package
		phpcount := 27 * archaius.Conf.Population / 100
		pname := "php"
		ppkg := "karyon" // karyon randomly calls its dependencies which isn't really right for master/slave/cache
		for i := r * phpcount; i < (r+1)*phpcount; i++ {
			phpname := names.Make(arch, rname, znames[i%2], pname, ppkg, i)
			noodles[phpname] = make(chan gotocol.Message)
			go karyon.Start(noodles[phpname])
			noodles[phpname] <- gotocol.Message{gotocol.Hello, listener, time.Now(), phpname}
			noodles[phpname] <- gotocol.Message{gotocol.Inform, eurekachan, time.Now(), ""}
			// connect all the php to mysql and memcached in one zone only
			noodles[phpname] <- gotocol.Message{gotocol.NameDrop, noodles[master], time.Now(), master}
			noodles[phpname] <- gotocol.Message{gotocol.NameDrop, noodles[slave], time.Now(), slave}
			noodles[phpname] <- gotocol.Message{gotocol.NameDrop, noodles[memname], time.Now(), memname}
			noodles[elbname] <- gotocol.Message{gotocol.NameDrop, noodles[phpname], time.Now(), phpname}
		}
	}
	// stop here for 5 for single region, then add second region for step 6
	if archaius.Conf.StopStep == 5 || archaius.Conf.StopStep == 6 {
		run(dnsname)
		return
	}
	run(dnsname)
}

// Run lamp for a while then shut down
func run(root string) {
	var msg gotocol.Message
	// tell denominator to start chatting with microservices every 0.01 secs
	delay := fmt.Sprintf("%dms", 10)
	log.Println("lamp: denominator activity rate ", delay)
	noodles[root] <- gotocol.Message{gotocol.Chat, nil, time.Now(), delay}

	// wait until the delay has finished
	if archaius.Conf.RunDuration >= time.Millisecond {
		time.Sleep(archaius.Conf.RunDuration)
	}
	log.Println("lamp: Shutdown")
	for _, noodle := range noodles {
		gotocol.Message{gotocol.Goodbye, nil, time.Now(), "shutdown"}.GoSend(noodle)
	}
	// listen one extra time to catch eureka reply
	gotocol.Message{gotocol.Goodbye, listener, time.Now(), "shutdown"}.GoSend(eurekachan)
	for len(noodles) > 0 {
		msg = <-listener
		if archaius.Conf.Msglog {
			log.Printf("lamp: %v\n", msg)
		}
		switch msg.Imposition {
		case gotocol.Goodbye:
			delete(noodles, msg.Intention)
			if archaius.Conf.Msglog {
				log.Printf("lamp: lamp %v shutdown, population: %v    \n", msg.Intention, len(noodles))
			}
		}
	}
	// wait for eureka to flush messages and exit
	eureka.Wg.Wait()
	collect.Save()
	log.Println("lamp: Exit")
}
