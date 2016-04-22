// Package main for spigo - simulate protocol interactions in go.
// Terminology is a mix of NetflixOSS, promise theory and flying spaghetti monster lore
package main

import (
	"flag"
	"github.com/adrianco/spigo/actors/edda"          // log configuration state
	"github.com/adrianco/spigo/tooling/archaius"     // store the config for global lookup
	"github.com/adrianco/spigo/tooling/architecture" // run an architecture from a json definition
	"github.com/adrianco/spigo/tooling/asgard"       // tools to create an architecture
	"github.com/adrianco/spigo/tooling/collect"      // metrics to extvar
	"github.com/adrianco/spigo/tooling/flow"         // flow logging
	"github.com/adrianco/spigo/tooling/fsm"          // fsm and pirates
	"github.com/adrianco/spigo/tooling/gotocol"      // message protocol spec
	"github.com/adrianco/spigo/tooling/migration"    // migration from LAMP to netflixoss
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

var reload, graphmlEnabled, graphjsonEnabled bool
var duration, cpucount int

// main handles command line flags and starts up an architecture
func main() {
	flag.StringVar(&archaius.Conf.Arch, "a", "netflixoss", "Architecture to create or read, fsm, migration, or read from json_arch/<arch>_arch.json")
	flag.IntVar(&archaius.Conf.Population, "p", 100, "Pirate population for fsm or scale factor % for other architectures")
	flag.IntVar(&duration, "d", 10, "Simulation duration in seconds")
	flag.IntVar(&archaius.Conf.Regions, "w", 1, "Wide area regions to replicate architecture into, defaults based on 6 AWS region names")
	flag.BoolVar(&graphmlEnabled, "g", false, "Enable GraphML logging of nodes and edges to gml/<arch>.graphml")
	flag.BoolVar(&graphjsonEnabled, "j", false, "Enable GraphJSON logging of nodes and edges to json/<arch>.json")
	flag.BoolVar(&archaius.Conf.Msglog, "m", false, "Enable console logging of every message")
	flag.BoolVar(&reload, "r", false, "Reload graph from json/<arch>.json to setup architecture")
	flag.BoolVar(&archaius.Conf.Collect, "c", false, "Collect to json_metrics csv_metricsand via http: extvars")
	flag.IntVar(&archaius.Conf.StopStep, "s", 0, "Sequence number to create multiple runs for ui to step through in json/<arch><s>.json")
	flag.StringVar(&archaius.Conf.EurekaPoll, "u", "1s", "Polling interval for Eureka name service, increase for large populations")
	flag.StringVar(&archaius.Conf.Keyvals, "kv", "", "Configuration key:value - chat:10ms sets default message insert rate")
	flag.BoolVar(&archaius.Conf.Filter, "f", false, "Filter output names to simplify graph by collapsing instances to services")
	flag.IntVar(&cpucount, "cpus", runtime.NumCPU(), "Number of CPUs for Go runtime")
	runtime.GOMAXPROCS(cpucount)
	var cpuprofile = flag.String("cpuprofile", "", "Write cpu profile to file")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if archaius.Conf.Collect {
		collect.Serve(8123) // start web server at port
	}
	if graphjsonEnabled || graphmlEnabled {
		if graphjsonEnabled {
			archaius.Conf.GraphjsonFile = archaius.Conf.Arch
		}
		if graphmlEnabled {
			archaius.Conf.GraphmlFile = archaius.Conf.Arch
		}
		// make a buffered channel so logging can start before edda is scheduled
		edda.Logchan = make(chan gotocol.Message, 100)
	}
	archaius.Conf.RunDuration = time.Duration(duration) * time.Second
	// start up the selected architecture
	go edda.Start(archaius.Conf.Arch + ".edda") // start edda first
	if reload {
		asgard.Run(asgard.Reload(archaius.Conf.Arch), "")
	} else {
		switch archaius.Conf.Arch {
		case "fsm":
			fsm.Start()
		case "migration":
			migration.Start() // step by step from lamp to netflixoss
		default:
			a := architecture.ReadArch(archaius.Conf.Arch)
			if a == nil {
				log.Fatal("Architecture " + archaius.Conf.Arch + " isn't recognized")
			} else {
				architecture.Start(a)
			}
		}
	}
	log.Println("spigo: complete")
	// stop edda if it's running and wait for edda to flush messages
	if edda.Logchan != nil {
		close(edda.Logchan)
	}
	edda.Wg.Wait()
	flow.Shutdown()
}
