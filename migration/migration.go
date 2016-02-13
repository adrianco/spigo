// Package migration implements a simulation of migration to a global large scale microservice architecture
// It creates and controls a collection of aws, lamp, netflixoss and netflix application microservices
// or reads in a network from a json file. It also logs the architecture (nodes and links) as it evolves
package migration

import (
	"github.com/adrianco/spigo/archaius"       // global configuration
	"github.com/adrianco/spigo/asgard"         // tools to create an architecture
	. "github.com/adrianco/spigo/packagenames" // name definitions
	"log"
)

// Start lamp to netflixoss step by step migration
func Start() {
	regions := archaius.Conf.Regions
	if archaius.Conf.Population < 1 {
		log.Fatal("migration: can't create less than 1 microservice")
	} else {
		log.Printf("migration: scaling to %v%%", archaius.Conf.Population)
	}
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
	case 0: // basic LAMP
		asgard.CreateChannels()
		asgard.CreateEureka() // service registries for each zone
		asgard.Create(sname, StorePkg, regions, mysqlcount, sname)
		asgard.Create(pname, MonolithPkg, regions, phpcount, sname)
		asgard.Create(elbname, ElbPkg, regions, 0, pname)
	case 1: // basic LAMP with memcache
		asgard.CreateChannels()
		asgard.CreateEureka() // service registries for each zone
		asgard.Create(sname, StorePkg, regions, mysqlcount, sname)
		asgard.Create(mname, StorePkg, regions, mcount)
		asgard.Create(pname, MonolithPkg, regions, phpcount, sname, mname)
		asgard.Create(elbname, ElbPkg, regions, 0, pname)
	case 2: // LAMP with zuul and memcache
		asgard.CreateChannels()
		asgard.CreateEureka() // service registries for each zone
		asgard.Create(sname, StorePkg, regions, mysqlcount, sname)
		asgard.Create(mname, StorePkg, regions, mcount)
		asgard.Create(pname, MonolithPkg, regions, phpcount, sname, mname)
		asgard.Create(zuname, ZuulPkg, regions, zuulcount, pname)
		asgard.Create(elbname, ElbPkg, regions, 0, zuname)
	case 3: // LAMP with zuul and staash and evcache
		asgard.CreateChannels()
		asgard.CreateEureka() // service registries for each zone
		asgard.Create(sname, StorePkg, regions, mysqlcount, sname)
		asgard.Create(mname, StorePkg, regions, mcount)
		asgard.Create(tname, StaashPkg, regions, staashcount, sname, mname)
		asgard.Create(pname, KaryonPkg, regions, phpcount, tname)
		asgard.Create(zuname, ZuulPkg, regions, zuulcount, pname)
		asgard.Create(elbname, ElbPkg, regions, 0, zuname)
	case 4: // added node microservice
		asgard.CreateChannels()
		asgard.CreateEureka() // service registries for each zone
		asgard.Create(sname, StorePkg, regions, mysqlcount, sname)
		asgard.Create(mname, StorePkg, regions, mcount)
		asgard.Create(tname, StaashPkg, regions, staashcount, sname, mname, cname)
		asgard.Create(pname, KaryonPkg, regions, phpcount, tname)
		asgard.Create(nname, KaryonPkg, regions, nodecount, tname)
		asgard.Create(zuname, ZuulPkg, regions, zuulcount, pname, nname)
		asgard.Create(elbname, ElbPkg, regions, 0, zuname)
	case 5: // added cassandra alongside mysql
		asgard.CreateChannels()
		asgard.CreateEureka() // service registries for each zone
		asgard.Create(cname, PriamCassandraPkg, regions, priamCassandracount, cname)
		asgard.Create(sname, StorePkg, regions, mysqlcount, sname)
		asgard.Create(mname, StorePkg, regions, mcount)
		asgard.Create(tname, StaashPkg, regions, staashcount, sname, mname, cname)
		asgard.Create(pname, KaryonPkg, regions, phpcount, tname)
		asgard.Create(nname, KaryonPkg, regions, nodecount, tname)
		asgard.Create(zuname, ZuulPkg, regions, zuulcount, pname, nname)
		asgard.Create(elbname, ElbPkg, regions, 0, zuname)
	case 6: // removed mysql so that multi-region will work properly
		asgard.CreateChannels()
		asgard.CreateEureka() // service registries for each zone
		asgard.Create(cname, PriamCassandraPkg, regions, priamCassandracount, cname)
		asgard.Create(mname, StorePkg, regions, mcount)
		asgard.Create(tname, StaashPkg, regions, staashcount, mname, cname)
		asgard.Create(pname, KaryonPkg, regions, phpcount, tname)
		asgard.Create(nname, KaryonPkg, regions, nodecount, tname)
		asgard.Create(zuname, ZuulPkg, regions, zuulcount, pname, nname)
		asgard.Create(elbname, ElbPkg, regions, 0, zuname)
	case 7: // set two regions with disconnected priamCassandra
		regions = 2
		archaius.Conf.Regions = regions
		asgard.CreateChannels()
		asgard.CreateEureka() // service registries for each zone
		asgard.Create(cname, PriamCassandraPkg, regions, priamCassandracount, cname)
		asgard.Create(mname, StorePkg, regions, mcount)
		asgard.Create(tname, StaashPkg, regions, staashcount, mname, cname)
		asgard.Create(pname, KaryonPkg, regions, phpcount, tname)
		asgard.Create(nname, KaryonPkg, regions, nodecount, tname)
		asgard.Create(zuname, ZuulPkg, regions, zuulcount, pname, nname)
		asgard.Create(elbname, ElbPkg, regions, 0, zuname)
	case 8: // set two regions with connected priamCassandra
		regions = 2
		archaius.Conf.Regions = regions
		asgard.CreateChannels()
		asgard.CreateEureka() // service registries for each zone
		asgard.Create(cname, PriamCassandraPkg, regions, priamCassandracount, "eureka", cname)
		asgard.Create(mname, StorePkg, regions, mcount)
		asgard.Create(tname, StaashPkg, regions, staashcount, mname, cname)
		asgard.Create(pname, KaryonPkg, regions, phpcount, tname)
		asgard.Create(nname, KaryonPkg, regions, nodecount, tname)
		asgard.Create(zuname, ZuulPkg, regions, zuulcount, pname, nname)
		asgard.Create(elbname, ElbPkg, regions, 0, zuname)
	case 9: // set three regions with disconnected priamCassandra
		regions = 3
		archaius.Conf.Regions = regions
		asgard.CreateChannels()
		asgard.CreateEureka() // service registries for each zone
		asgard.Create(cname, PriamCassandraPkg, regions, priamCassandracount, "eureka", cname)
		asgard.Create(mname, StorePkg, regions, mcount)
		asgard.Create(tname, StaashPkg, regions, staashcount, mname, cname)
		asgard.Create(pname, KaryonPkg, regions, phpcount, tname)
		asgard.Create(nname, KaryonPkg, regions, nodecount, tname)
		asgard.Create(zuname, ZuulPkg, regions, zuulcount, pname, nname)
		asgard.Create(elbname, ElbPkg, regions, 0, zuname)
	}
	dnsname := asgard.Create(dns, DenominatorPkg, 0, 0, elbname)
	asgard.Run(dnsname, "")
}
