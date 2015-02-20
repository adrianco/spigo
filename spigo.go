// Package main for spigo - simulate protocol interactions in go.
// Terminology is a mix of promise theory and flying spaghetti monster lore
package main

import (
	"flag"
	"github.com/adrianco/spigo/archaius"   // store the config for global lookup
	"github.com/adrianco/spigo/collect"    // metrics to extvar
	"github.com/adrianco/spigo/edda"       // log configuration state
	"github.com/adrianco/spigo/fsm"        // fsm and pirates
	"github.com/adrianco/spigo/gotocol"    // message protocol spec
	"github.com/adrianco/spigo/netflixoss" // start the netflix opensource microservices
	"log"
	"os"
	"runtime/pprof"
	"time"
)

var reload, graphmlEnabled, graphjsonEnabled bool
var duration int

// main handles command line flags and starts up an architecture
func main() {
	flag.StringVar(&archaius.Conf.Arch, "a", "fsm", "Architecture to create or read, fsm or netflixoss")
	flag.IntVar(&archaius.Conf.Population, "p", 100, "  Pirate population for fsm or scale factor % for netflixoss")
	flag.IntVar(&duration, "d", 10, "   Simulation duration in seconds")
	flag.IntVar(&archaius.Conf.Regions, "w", 1, "    Wide area regions")
	flag.BoolVar(&graphmlEnabled, "g", false, "Enable GraphML logging of nodes and edges to <arch>.graphml")
	flag.BoolVar(&graphjsonEnabled, "j", false, "Enable GraphJSON logging of nodes and edges to <arch>.json")
	flag.BoolVar(&archaius.Conf.Msglog, "m", false, "Enable console logging of every message")
	flag.BoolVar(&reload, "r", false, "Reload <arch>.json to setup architecture")
	flag.BoolVar(&archaius.Conf.Collect, "c", false, "Collect metrics to <arch>_metrics.json and via http:")
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
	switch archaius.Conf.Arch {
	case "fsm":
		go edda.Start("fsm.edda") // start edda first
		if reload {
			fsm.Reload(archaius.Conf.Arch)
		} else {
			fsm.Start()
		}
		log.Println("spigo: fsm complete")
	case "netflixoss":
		go edda.Start("netflixoss.edda") // start edda first
		if reload {
			netflixoss.Reload(archaius.Conf.Arch)
		} else {
			netflixoss.Start()
		}
		log.Println("spigo: netflixoss complete")
	default:
		log.Fatal("Architecture " + archaius.Conf.Arch + " isn't recognized")
	}
	// stop edda if it's running and wait for edda to flush messages
	if edda.Logchan != nil {
		close(edda.Logchan)
	}
	edda.Wg.Wait()
}
