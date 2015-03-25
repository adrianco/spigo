// Package migration implements a simulation of migration to a global large scale microservice architecture
// It creates and controls a collection of aws, lamp, netflixoss and netflix application microservices
// or reads in a network from a json file. It also logs the architecture (nodes and links) as it evolves
package migration

import (
	"fmt"
	"github.com/adrianco/spigo/archaius" // global configuration
	"github.com/adrianco/spigo/asgard"   // tools to create an architecture
	"github.com/adrianco/spigo/collect"  // metrics collector
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/names" // manage service name hierarchy
	"log"
	"time"
)

// Reload the network from a file
func Reload(arch string) {
	root := ""
	log.Println("migration reloading from " + arch + ".json")
	g := graphjson.ReadArch(arch)
	archaius.Conf.Population = 0 // just to make sure
	// count how many nodes there are
	for _, element := range g.Graph {
		if element.Node != "" {
			archaius.Conf.Population++
		}
	}
	asgard.CreateChannels()
	asgard.CreateEureka()
	// eureka and edda aren't recorded in the json file to simplify the graph
	// Start all the services
	for _, element := range g.Graph {
		if element.Node != "" {
			name := element.Node
			asgard.StartNode(name, nil)
			if names.Package(name) == asgard.DenominatorPkg {
				root = name
			}
		}
	}
	// Make all the connections
	for _, element := range g.Graph {
		if element.Edge != "" && element.Source != "" && element.Target != "" {
			asgard.Connect(element.Source, element.Target)
		}
	}
	// run for a while
	if root == "" {
		log.Fatal("No denominator root microservice specified")
	}
	run(root)
}

// Start lamp to netflixoss step by step migration
func Start() {
	arch := archaius.Conf.Arch
	rnames := archaius.Conf.RegionNames
	znames := archaius.Conf.ZoneNames
	if archaius.Conf.Population < 1 {
		log.Fatal("migration: can't create less than 1 microservice")
	} else {
		log.Printf("migration: scaling to %v%%", archaius.Conf.Population)
	}
	asgard.CreateChannels()
	asgard.CreateEureka() // service registries for each zone
	// Build the configuration step by step
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
	switch archaius.Conf.StopStep {
	case 1: // basic LAMP with memcache
		asgard.Create(sname, asgard.StorePkg, archaius.Conf.Regions, mysqlcount, sname)
		asgard.Create(mname, asgard.StorePkg, archaius.Conf.Regions, mcount)
		asgard.Create(pname, asgard.MonolithPkg, archaius.Conf.Regions, phpcount, sname, mname)
		asgard.Create(elbname, asgard.ElbPkg, archaius.Conf.Regions, 0, pname)
	case 2: // LAMP with zuul and memcache
		asgard.Create(sname, asgard.StorePkg, archaius.Conf.Regions, mysqlcount, sname)
		asgard.Create(mname, asgard.StorePkg, archaius.Conf.Regions, mcount)
		asgard.Create(pname, asgard.MonolithPkg, archaius.Conf.Regions, phpcount, sname, mname)
		asgard.Create(zuname, asgard.ZuulPkg, archaius.Conf.Regions, zuulcount, pname)
		asgard.Create(elbname, asgard.ElbPkg, archaius.Conf.Regions, 0, zuname)
	case 3: // LAMP with zuul and staash and evcache
		asgard.Create(sname, asgard.StorePkg, archaius.Conf.Regions, mysqlcount, sname)
		asgard.Create(mname, asgard.StorePkg, archaius.Conf.Regions, mcount)
		asgard.Create(tname, asgard.StaashPkg, archaius.Conf.Regions, staashcount, sname, mname)
		asgard.Create(pname, asgard.KaryonPkg, archaius.Conf.Regions, phpcount, tname)
		asgard.Create(zuname, asgard.ZuulPkg, archaius.Conf.Regions, zuulcount, pname)
		asgard.Create(elbname, asgard.ElbPkg, archaius.Conf.Regions, 0, zuname)
	case 4: // added node microservice
		asgard.Create(sname, asgard.StorePkg, archaius.Conf.Regions, mysqlcount, sname)
		asgard.Create(mname, asgard.StorePkg, archaius.Conf.Regions, mcount)
		asgard.Create(tname, asgard.StaashPkg, archaius.Conf.Regions, staashcount, sname, mname, cname)
		asgard.Create(pname, asgard.KaryonPkg, archaius.Conf.Regions, phpcount, tname)
		asgard.Create(nname, asgard.KaryonPkg, archaius.Conf.Regions, nodecount, tname)
		asgard.Create(zuname, asgard.ZuulPkg, archaius.Conf.Regions, zuulcount, pname, nname)
		asgard.Create(elbname, asgard.ElbPkg, archaius.Conf.Regions, 0, zuname)
	case 5: // added cassandra alongside mysql
		asgard.Create(cname, "priamCassandra", archaius.Conf.Regions, priamCassandracount, cname)
		asgard.Create(sname, asgard.StorePkg, archaius.Conf.Regions, mysqlcount, sname)
		asgard.Create(mname, asgard.StorePkg, archaius.Conf.Regions, mcount)
		asgard.Create(tname, asgard.StaashPkg, archaius.Conf.Regions, staashcount, sname, mname, cname)
		asgard.Create(pname, asgard.KaryonPkg, archaius.Conf.Regions, phpcount, tname)
		asgard.Create(nname, asgard.KaryonPkg, archaius.Conf.Regions, nodecount, tname)
		asgard.Create(zuname, asgard.ZuulPkg, archaius.Conf.Regions, zuulcount, pname, nname)
		asgard.Create(elbname, asgard.ElbPkg, archaius.Conf.Regions, 0, zuname)
	default: // for all higher steps
		fallthrough
	case 6: // removed mysql so that multi-region will work properly
		asgard.Create(cname, asgard.PriamCassandraPkg, archaius.Conf.Regions, priamCassandracount, cname)
		asgard.Create(mname, asgard.StorePkg, archaius.Conf.Regions, mcount)
		asgard.Create(tname, asgard.StaashPkg, archaius.Conf.Regions, staashcount, mname, cname)
		asgard.Create(pname, asgard.KaryonPkg, archaius.Conf.Regions, phpcount, tname)
		asgard.Create(nname, asgard.KaryonPkg, archaius.Conf.Regions, nodecount, tname)
		asgard.Create(zuname, asgard.ZuulPkg, archaius.Conf.Regions, zuulcount, pname, nname)
		asgard.Create(elbname, asgard.ElbPkg, archaius.Conf.Regions, 0, zuname)
	}
	dnsname := asgard.Create(dns, asgard.DenominatorPkg, 0, 0, elbname)
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
				pC := names.Make(arch, rnames[r], znames[i%3], cname, asgard.PriamCassandraPkg, i)
				// for each of the other regions connect to one node
				for j := 1; j < archaius.Conf.Regions; j++ {
					pCindex := (i + j*priamCassandracount) % (archaius.Conf.Regions * priamCassandracount)
					pCremote := names.Make(arch, rnames[(r+1)%archaius.Conf.Regions], znames[pCindex%3], cname, asgard.PriamCassandraPkg, pCindex)
					asgard.Connect(pC, pCremote)
				}
			}
		}
	}
	run(dnsname)
}

// Run migration for a while then shut down
func run(rootservice string) {
	// tell denominator to start chatting with microservices every 0.01 secs
	delay := fmt.Sprintf("%dms", 10)
	log.Println("migration: denominator activity rate ", delay)
	asgard.SendToName(rootservice, gotocol.Message{gotocol.Chat, nil, time.Now(), delay})

	// wait until the delay has finished
	if archaius.Conf.RunDuration >= time.Millisecond {
		time.Sleep(archaius.Conf.RunDuration)
	}
	log.Println("migration: Shutdown")
	asgard.ShutdownNodes()
	asgard.ShutdownEureka()
	collect.Save()
	log.Println("migration: Exit")
}
