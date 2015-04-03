// Package migration implements a simulation of migration to a global large scale microservice architecture
// It creates and controls a collection of aws, lamp, netflixoss and netflix application microservices
// or reads in a network from a json file. It also logs the architecture (nodes and links) as it evolves
package migration

import (
	"github.com/adrianco/spigo/archaius" // global configuration
	"github.com/adrianco/spigo/asgard"   // tools to create an architecture
	"github.com/adrianco/spigo/names"    // manage service name hierarchy
	"log"
)

// Start lamp to netflixoss step by step migration
func Start() {
	// make some shorter names
	arch := archaius.Conf.Arch
	rnames := archaius.Conf.RegionNames
	znames := archaius.Conf.ZoneNames
	regions := archaius.Conf.Regions
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
		asgard.Create(sname, asgard.StorePkg, regions, mysqlcount, sname)
		asgard.Create(mname, asgard.StorePkg, regions, mcount)
		asgard.Create(pname, asgard.MonolithPkg, regions, phpcount, sname, mname)
		asgard.Create(elbname, asgard.ElbPkg, regions, 0, pname)
	case 2: // LAMP with zuul and memcache
		asgard.Create(sname, asgard.StorePkg, regions, mysqlcount, sname)
		asgard.Create(mname, asgard.StorePkg, regions, mcount)
		asgard.Create(pname, asgard.MonolithPkg, regions, phpcount, sname, mname)
		asgard.Create(zuname, asgard.ZuulPkg, regions, zuulcount, pname)
		asgard.Create(elbname, asgard.ElbPkg, regions, 0, zuname)
	case 3: // LAMP with zuul and staash and evcache
		asgard.Create(sname, asgard.StorePkg, regions, mysqlcount, sname)
		asgard.Create(mname, asgard.StorePkg, regions, mcount)
		asgard.Create(tname, asgard.StaashPkg, regions, staashcount, sname, mname)
		asgard.Create(pname, asgard.KaryonPkg, regions, phpcount, tname)
		asgard.Create(zuname, asgard.ZuulPkg, regions, zuulcount, pname)
		asgard.Create(elbname, asgard.ElbPkg, regions, 0, zuname)
	case 4: // added node microservice
		asgard.Create(sname, asgard.StorePkg, regions, mysqlcount, sname)
		asgard.Create(mname, asgard.StorePkg, regions, mcount)
		asgard.Create(tname, asgard.StaashPkg, regions, staashcount, sname, mname, cname)
		asgard.Create(pname, asgard.KaryonPkg, regions, phpcount, tname)
		asgard.Create(nname, asgard.KaryonPkg, regions, nodecount, tname)
		asgard.Create(zuname, asgard.ZuulPkg, regions, zuulcount, pname, nname)
		asgard.Create(elbname, asgard.ElbPkg, regions, 0, zuname)
	case 5: // added cassandra alongside mysql
		asgard.Create(cname, asgard.PriamCassandraPkg, regions, priamCassandracount, cname)
		asgard.Create(sname, asgard.StorePkg, regions, mysqlcount, sname)
		asgard.Create(mname, asgard.StorePkg, regions, mcount)
		asgard.Create(tname, asgard.StaashPkg, regions, staashcount, sname, mname, cname)
		asgard.Create(pname, asgard.KaryonPkg, regions, phpcount, tname)
		asgard.Create(nname, asgard.KaryonPkg, regions, nodecount, tname)
		asgard.Create(zuname, asgard.ZuulPkg, regions, zuulcount, pname, nname)
		asgard.Create(elbname, asgard.ElbPkg, regions, 0, zuname)
	default: // for all higher steps
		fallthrough
	case 6: // removed mysql so that multi-region will work properly
		asgard.Create(cname, asgard.PriamCassandraPkg, regions, priamCassandracount, cname)
		asgard.Create(mname, asgard.StorePkg, regions, mcount)
		asgard.Create(tname, asgard.StaashPkg, regions, staashcount, mname, cname)
		asgard.Create(pname, asgard.KaryonPkg, regions, phpcount, tname)
		asgard.Create(nname, asgard.KaryonPkg, regions, nodecount, tname)
		asgard.Create(zuname, asgard.ZuulPkg, regions, zuulcount, pname, nname)
		asgard.Create(elbname, asgard.ElbPkg, regions, 0, zuname)
	}
	dnsname := asgard.Create(dns, asgard.DenominatorPkg, 0, 0, elbname)
	// stop here for for single region, then add second region, then join them
	if archaius.Conf.StopStep < 8 {
		asgard.Run(dnsname)
		return
	}
	// Connect cross region Cassandra0
	if regions > 1 {
		// for each region
		for r := 0; r < regions; r++ {
			// for each priamCassandrian in that region
			for i := r * priamCassandracount; i < (r+1)*priamCassandracount; i++ {
				pC := names.Make(arch, rnames[r], znames[i%3], cname, asgard.PriamCassandraPkg, i)
				// for each of the other regions connect to one node
				for j := 1; j < regions; j++ {
					pCindex := (i + j*priamCassandracount) % (regions * priamCassandracount)
					pCremote := names.Make(arch, rnames[(r+1)%regions], znames[pCindex%3], cname, asgard.PriamCassandraPkg, pCindex)
					asgard.Connect(pC, pCremote)
				}
			}
		}
	}
	asgard.Run(dnsname)
}
