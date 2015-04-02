// Package netflixoss implements a global large scale microservice architecture
// It creates and controls a collection of aws and netflix application microservices
// or reads in a network from a json file. It also logs the architecture (nodes and links) as it evolves
package netflixoss

import (
	"fmt"
	"github.com/adrianco/spigo/archaius" // global configuration
	"github.com/adrianco/spigo/asgard"   // tools to create an architecture
	"github.com/adrianco/spigo/collect"  // metrics collector
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/names" // manage service name hierarchy
	"log"
	"time"
)

// Reload the network from a file
func Reload(arch string) {
	run(asgard.Reload(arch))
}

// Start netflixoss
func Start() {
	arch := archaius.Conf.Arch
	rnames := archaius.Conf.RegionNames
	znames := archaius.Conf.ZoneNames
	if archaius.Conf.Population < 1 {
		log.Fatal("netflixoss: can't create less than 1 microservice")
	} else {
		log.Printf("netflixoss: scaling to %v%%", archaius.Conf.Population)
	}
	asgard.CreateChannels()
	asgard.CreateEureka() // service registries for each zone

	// priam managed Cassandra cluster, turtle because it's used to configure other clusters
	priamCassandracount := 12 * archaius.Conf.Population / 100
	cname := "cassTurtle"
	// staash data access layer connects to mysql master and slave, and evcache
	staashcount := 6 * archaius.Conf.Population / 100
	tname := "turtle"
	// some node microservice logic, we can create a network of simple services from the karyon package
	nodecount := 24 * archaius.Conf.Population / 100
	nname := "node"
	// some java microservice logic, we can create a network of simple services from the karyon package
	javacount := 18 * archaius.Conf.Population / 100
	jname := "javaweb" // zuul api proxies
	zuulcount := 9 * archaius.Conf.Population / 100
	zuname := "wwwproxy"
	// AWS elastic load balancer
	elbname := "www-elb"
	// DNS endpoint
	dns := "www"

	asgard.Create(cname, asgard.PriamCassandraPkg, archaius.Conf.Regions, priamCassandracount, cname)
	asgard.Create(tname, asgard.StaashPkg, archaius.Conf.Regions, staashcount, cname)
	asgard.Create(jname, asgard.KaryonPkg, archaius.Conf.Regions, javacount, tname)
	asgard.Create(nname, asgard.KaryonPkg, archaius.Conf.Regions, nodecount, jname)
	asgard.Create(zuname, asgard.ZuulPkg, archaius.Conf.Regions, zuulcount, nname)
	asgard.Create(elbname, asgard.ElbPkg, archaius.Conf.Regions, 0, zuname)

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

// Run netflixoss for a while then shut down
func run(rootservice string) {
	// tell denominator to start chatting with microservices every 0.01 secs
	delay := fmt.Sprintf("%dms", 10)
	log.Println("netflixoss: denominator activity rate ", delay)
	asgard.SendToName(rootservice, gotocol.Message{gotocol.Chat, nil, time.Now(), delay})

	// wait until the delay has finished
	if archaius.Conf.RunDuration >= time.Millisecond {
		time.Sleep(archaius.Conf.RunDuration)
	}
	log.Println("netflixoss: Shutdown")
	asgard.ShutdownNodes()
	asgard.ShutdownEureka()
	collect.Save()
	log.Println("netflixoss: Exit")
}
