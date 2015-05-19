// Package architecture reads a microservice architecture definition from a file
// It creates and controls a collection of aws and netflix application microservices
package netflixoss

import (
	"github.com/adrianco/spigo/archaius" // global configuration
	"github.com/adrianco/spigo/asgard"   // tools to create an architecture
	"log"
)


// Start architecture
func Start() {
	//regions := archaius.Conf.Regions
	if archaius.Conf.Population < 1 {
		log.Fatal("architecture: can't create less than 1 microservice")
	} else {
		log.Printf("architecture: scaling to %v%%", archaius.Conf.Population)
	}
	asgard.CreateChannels()
	asgard.CreateEureka() // service registries for each zone


//	asgard.Create(cname, asgard.PriamCassandraPkg, regions, priamCassandracount, "eureka", cname)
	asgard.Run(asgard.Create("www", asgard.DenominatorPkg, 0, 0, ""), "")
}
