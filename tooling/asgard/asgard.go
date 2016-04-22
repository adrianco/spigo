// Package asgard contains shared code that is used to create an aws/netflixoss style architecture (including lamp and migration)
package asgard

import (
	"fmt"
	"github.com/adrianco/spigo/actors/denominator"    // DNS service
	"github.com/adrianco/spigo/actors/elb"            // elastic load balancer
	"github.com/adrianco/spigo/actors/eureka"         // service and attribute registry
	"github.com/adrianco/spigo/actors/karyon"         // business logic microservice
	"github.com/adrianco/spigo/actors/monolith"       // business logic monolith
	. "github.com/adrianco/spigo/actors/packagenames" // name definitions
	"github.com/adrianco/spigo/actors/pirate"         // random end user network
	"github.com/adrianco/spigo/actors/priamCassandra" // Priam managed Cassandra cluster
	"github.com/adrianco/spigo/actors/staash"         // storage tier as a service http - data access layer
	"github.com/adrianco/spigo/actors/store"          // generic storage service
	"github.com/adrianco/spigo/actors/zuul"           // API proxy microservice router
	"github.com/adrianco/spigo/tooling/archaius"      // global configuration
	"github.com/adrianco/spigo/tooling/chaosmonkey"   // delete nodes at random
	"github.com/adrianco/spigo/tooling/collect"       // metrics collector
	"github.com/adrianco/spigo/tooling/gotocol"
	"github.com/adrianco/spigo/tooling/graphjson"
	"github.com/adrianco/spigo/tooling/names" // manage service name hierarchy
	"log"
	"time"
)

var (
	listener   chan gotocol.Message            // asgard listener
	eurekachan map[string]chan gotocol.Message // eureka for each region and zone
	// noodles channels mapped by microservice name connects netflixoss to everyone
	noodles map[string]chan gotocol.Message
)

// Create the maps of channels
func CreateChannels() {
	listener = make(chan gotocol.Message) // listener for architecture
	noodles = make(map[string]chan gotocol.Message, archaius.Conf.Population)
	eurekachan = make(map[string]chan gotocol.Message, 3*archaius.Conf.Regions)
}

type mapchan map[string]chan gotocol.Message

// Create a tier and specify any dependencies, return the full name of the last node created
func Create(servicename, packagename string, regions, count int, dependencies ...string) string {
	var name string
	arch := archaius.Conf.Arch
	rnames := archaius.Conf.RegionNames
	znames := archaius.Conf.ZoneNames
	if regions == 0 { // for dns that isn't in a region or zone
		//log.Printf("Create cross region: " + servicename)
		name = names.Make(arch, "*", "*", servicename, packagename, 0)
		StartNode(name, dependencies...)
	}
	for r := 0; r < regions; r++ {
		if count == 0 { // for AWS services that are cross zone like elb and S3
			//log.Printf("Create cross zone: " + servicename)
			name = names.Make(arch, rnames[r], "*", servicename, packagename, 0)
			StartNode(name, dependencies...)
		} else {
			//log.Printf("Create service: " + servicename)
			cass := make(map[string]mapchan) // for token distribution
			for i := r * count; i < (r+1)*count; i++ {
				name = names.Make(arch, rnames[r], znames[i%3], servicename, packagename, i)
				//log.Println(dependencies)
				StartNode(name, dependencies...)
				if packagename == "priamCassandra" {
					rz := names.RegionZone(name)
					if cass[rz] == nil {
						cass[rz] = make(mapchan)
					}
					cass[rz][name] = noodles[name] // remember the nodes
				}
			}
			if packagename == "priamCassandra" {
				// split by zone
				for _, v := range cass {
					priamCassandra.Distribute(v) // returns a string if it needs logging
				}
			}
		}
	}
	return name
}

// Reload the network from a file
func Reload(arch string) string {
	root := ""
	g := graphjson.ReadArch(arch)
	archaius.Conf.Population = 0 // just to make sure
	// count how many nodes there are
	for _, element := range g.Graph {
		if element.Node != "" {
			archaius.Conf.Population++
		}
	}
	CreateChannels()
	CreateEureka()
	// eureka and edda aren't recorded in the json file to simplify the graph
	// Start all the services
	cass := make(map[string]chan gotocol.Message) // for token distribution
	for _, element := range g.Graph {
		if element.Node != "" {
			name := element.Node
			StartNode(name, "")
			if names.Package(name) == DenominatorPkg {
				root = name
			}
			if names.Package(name) == "priamCassandra" {
				cass[name] = noodles[name] // remember the nodes
			}
		}
	}
	if len(cass) > 0 { // currently doesn't handle multiple priamCassandra per arch
		priamCassandra.Distribute(cass) // returns a string if it needs logging
	}
	// Make all the connections
	for _, element := range g.Graph {
		if element.Edge != "" && element.Source != "" && element.Target != "" {
			Connect(element.Source, element.Target)
		}
	}
	// run for a while
	if root == "" {
		log.Fatal("No denominator root microservice specified")
	}
	return root
}

// Tell a source node how to connect to a target node directly by name, only used when Eureka can't be used
func Connect(source, target string) {
	if noodles[source] != nil && noodles[target] != nil {
		gotocol.Send(noodles[source], gotocol.Message{gotocol.NameDrop, noodles[target], time.Now(), gotocol.NilContext, target})
		//log.Println("Link " + source + " > " + target)
	} else {
		log.Fatal("Asgard can't link " + source + " > " + target)
	}
}

// send a message directly to a name via asgard, only used during setup
func SendToName(name string, msg gotocol.Message) {
	if noodles[name] != nil {
		gotocol.Send(noodles[name], msg)
	} else {
		log.Fatal("Asgard can't send to " + name)
	}
}

// Start a node using the named package, and connect it to any dependencies
func StartNode(name string, dependencies ...string) {
	if names.Package(name) == EurekaPkg {
		eurekachan[name] = make(chan gotocol.Message, archaius.Conf.Population/3) // buffer sized to a zone
		go eureka.Start(eurekachan[name], name)
		return
	} else {
		noodles[name] = make(chan gotocol.Message)
	}
	// start the service and tell it it's name
	switch names.Package(name) {
	case PiratePkg:
		go pirate.Start(noodles[name])
	case ElbPkg:
		go elb.Start(noodles[name])
	case DenominatorPkg:
		go denominator.Start(noodles[name])
	case ZuulPkg:
		go zuul.Start(noodles[name])
	case KaryonPkg:
		go karyon.Start(noodles[name])
	case MonolithPkg:
		go monolith.Start(noodles[name])
	case StaashPkg:
		go staash.Start(noodles[name])
	case RiakPkg:
		fallthrough // fake Riak using priamCassandra
	case PriamCassandraPkg:
		go priamCassandra.Start(noodles[name])
	case CachePkg:
		fallthrough // fake memcache using store
	case VolumePkg:
		fallthrough // fake disk volume using store
	case StorePkg:
		go store.Start(noodles[name])
	default:
		log.Fatal("asgard: unknown package: " + names.Package(name))
	}
	noodles[name] <- gotocol.Message{gotocol.Hello, listener, time.Now(), gotocol.NilContext, name}
	// there is a eureka service registry in each zone, so in-zone services just get to talk to their local registry
	// elb are cross zone, so need to see all registries in a region
	// denominator are cross region so need to see all registries globally
	// priamCassandra depends explicitly on eureka for cross region clusters
	crossregion := false
	for _, d := range dependencies {
		if d == "eureka" {
			crossregion = true
		}
	}
	for n, ch := range eurekachan {
		if names.Region(name) == "*" || crossregion {
			// need to know every eureka in all zones and regions
			gotocol.Send(noodles[name], gotocol.Message{gotocol.Inform, ch, time.Now(), gotocol.NilContext, n})
		} else {
			if names.Zone(name) == "*" && names.Region(name) == names.Region(n) {
				// need every eureka in my region
				gotocol.Send(noodles[name], gotocol.Message{gotocol.Inform, ch, time.Now(), gotocol.NilContext, n})
			} else {
				if names.RegionZone(name) == names.RegionZone(n) {
					// just the eureka in this specific zone
					gotocol.Send(noodles[name], gotocol.Message{gotocol.Inform, ch, time.Now(), gotocol.NilContext, n})
				}
			}
		}
	}
	//log.Println(dependencies)
	// pass on symbolic dependencies without channels that will be looked up in Eureka later
	for _, dep := range dependencies {
		if dep != "" && dep != "eureka" { // ignore special case of eureka in dependency list
			//log.Println(name + " depends on " + dep)
			gotocol.Send(noodles[name], gotocol.Message{gotocol.NameDrop, nil, time.Now(), gotocol.NilContext, dep})
		}
	}
}

// create eureka service registries in each zone
func CreateEureka() {
	// setup name service and cross zone replication links
	znames := archaius.Conf.ZoneNames
	Create("eureka", EurekaPkg, archaius.Conf.Regions, 3)
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
				//log.Println("Eureka cross connect from: " + n + " to " + nn)
				gotocol.Send(ch, gotocol.Message{gotocol.NameDrop, cch, time.Now(), gotocol.NilContext, nn})
			}
		}
	}
}

// connect to eureka services in every region
func ConnectEveryEureka(name string) {
	for n, ch := range eurekachan {
		gotocol.Send(noodles[name], gotocol.Message{gotocol.Inform, ch, time.Now(), gotocol.NilContext, n})
	}
}

// Run architecture for a while then shut down
func Run(rootservice, victim string) {
	// tell denominator to start chatting with microservices every 0.01 secs by default
	delay := archaius.Key(archaius.Conf, "chat")
	if delay == "" {
		delay = fmt.Sprintf("%dms", 10)
	}
	log.Println(rootservice+" activity rate ", delay)
	SendToName(rootservice, gotocol.Message{gotocol.Chat, nil, time.Now(), gotocol.NilContext, delay})
	// wait until the delay has finished
	if archaius.Conf.RunDuration >= time.Millisecond {
		time.Sleep(archaius.Conf.RunDuration / 2)
		chaosmonkey.Delete(&noodles, victim) // kill a random victim half way through
		time.Sleep(archaius.Conf.RunDuration / 2)
	}
	log.Println("asgard: Shutdown")
	ShutdownNodes()
	ShutdownEureka()
	collect.Save()
}

// shut down the nodes and wait for them to go away
func ShutdownNodes() {
	for _, noodle := range noodles {
		gotocol.Message{gotocol.Goodbye, nil, time.Now(), gotocol.NilContext, "shutdown"}.GoSend(noodle)
	}
	for len(noodles) > 0 {
		msg := <-listener
		if archaius.Conf.Msglog {
			log.Printf("asgard: %v\n", msg)
		}
		switch msg.Imposition {
		case gotocol.Goodbye:
			delete(noodles, msg.Intention)
			if archaius.Conf.Msglog {
				log.Printf("asgard: %v shutdown, population: %v    \n", msg.Intention, len(noodles))
			}
		}
	}
}

// shut down the Eureka service registries and wait for them to go away
func ShutdownEureka() {
	// shutdown eureka and wait to catch eureka reply
	//log.Println(eurekachan)
	for _, ch := range eurekachan {
		gotocol.Message{gotocol.Goodbye, listener, time.Now(), gotocol.NilContext, "shutdown"}.GoSend(ch)
	}
	for _ = range eurekachan {
		<-listener
	}
	// wait for all the eureka to flush messages and exit
	eureka.Wg.Wait()
}
