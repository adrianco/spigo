// Package lamp implements a global large scale microservice architecture
// It creates and controls a collection of aws and netflix application microservices
// or reads in a network from a json file. It also logs the architecture (nodes and links) as it evolves
package lamp

import (
	"github.com/adrianco/spigo/archaius" // global configuration
	"github.com/adrianco/spigo/asgard"   // tools to create an architecture
	"log"
)

// Start lamp
func Start() {
	regions := archaius.Conf.Regions
	if archaius.Conf.Population < 1 {
		log.Fatal("lamp: can't create less than 1 microservice")
	} else {
		log.Printf("lamp: scaling to %v%%", archaius.Conf.Population)
	}
	asgard.CreateChannels()
	asgard.CreateEureka() // service registries for each zone
	// start mysql data store layer, which connects to itself
	mysqlcount := 2
	sname := "rds-mysql"
	// start memcached layer, only one per region
	mname := "memcache"
	mcount := 1
	// some php monolith logic, we can create a network of simple services from the karyon package
	phpcount := 18 * archaius.Conf.Population / 100
	pname := "phpweb"
	// AWS elastic load balancer
	elbname := "www-elb"
	// DNS endpoint
	dns := "www"
	asgard.Create(sname, asgard.StorePkg, regions, mysqlcount, sname)
	asgard.Create(mname, asgard.StorePkg, regions, mcount)
	asgard.Create(pname, asgard.MonolithPkg, regions, phpcount, sname, mname)
	asgard.Create(elbname, asgard.ElbPkg, regions, 0, pname)
	dnsname := asgard.Create(dns, asgard.DenominatorPkg, 0, 0, elbname)
	asgard.Run(dnsname, "")
}
