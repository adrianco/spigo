// Package netflixoss implements a global large scale microservice architecture
// It creates and controls a collection of aws and netflix application microservices
// or reads in a network from a json file. It also logs the architecture (nodes and links) as it evolves
package netflixoss

import (
	"github.com/adrianco/spigo/archaius" // global configuration
	"github.com/adrianco/spigo/asgard"   // tools to create an architecture
	"log"
)

// Start netflixoss
func Start() {
	regions := archaius.Conf.Regions
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

	asgard.Create(cname, asgard.PriamCassandraPkg, regions, priamCassandracount, "eureka", cname)
	asgard.Create(tname, asgard.StaashPkg, regions, staashcount, cname)
	asgard.Create(jname, asgard.KaryonPkg, regions, javacount, tname)
	asgard.Create(nname, asgard.KaryonPkg, regions, nodecount, jname)
	asgard.Create(zuname, asgard.ZuulPkg, regions, zuulcount, nname)
	asgard.Create(elbname, asgard.ElbPkg, regions, 0, zuname)
	asgard.Run(asgard.Create(dns, asgard.DenominatorPkg, 0, 0, elbname))
}
