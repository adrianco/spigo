spigo and simianviz
===================

[![Join the chat at https://gitter.im/adrianco/spigo](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/adrianco/spigo?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

The name spigo is taken, but simianviz wasn't, so domains have been registered etc. and the name will transition over the coming months.

[![GoDoc](https://godoc.org/github.com/adrianco/spigo?status.svg)](https://godoc.org/github.com/adrianco/spigo) [Wiki Instructions](https://github.com/adrianco/spigo/wiki)

Simulate Protocol Interactions in Go using nanoservice actors - spigo

SIMulate Interactive Actor Network VIsualiZation - simianviz - also visualize the simian army in action (not yet implemented).

Current work in progress adds context to each message for zipkin style request tracing, adds containers to the naming hierarchy and adds configurable filters to hide the extra levels when they aren't being used. The new -f option turns up filtering to produce a graph of services rather than nodes. Nodes also now have fake IP addresses. Docker compose yaml files can be converted to architecture json using
```
$ ./compose2arch -file compose.yaml > arch.json
```

Recent UI changes include pinning nodes so graphs can be stretched out and deleting nodes and edges by double-clicking on them. To add new entries to the architecture menu, add a line to ui/js/toolbar/index.js
[Run the netflixoss simulation in your browser](http://simianviz.surge.sh/netflixoss)

For a local installation of spigo, with no network dependencies, you can start the service and browse localhost:8000 using:
```
$ cd ui
$ npm install
$ npm run dev
```

Suitable for fairly large scale simulations, spigo runs well up to 100,000 independent nanoservice actors in a few GB of RAM. Three types of architectural models are implemented. One creates a peer to peer social network (fsm and pirates). Most others are based on a LAMP stack or NetflixOSS microservices in a more tree structured model loaded from an architecture definition file. The migration architecture is hard coded, starts with LAMP and ends with NetflixOSS.

Each nanoservice actor is a goroutine. to create 100,000 pirates, deliver 700,000 messages and wait to shut them all down again took about 4 seconds. The resulting graph can be visualized via GraphML or rendered by saving to Graph JSON and viewing in a web browser via D3.

A few lines of code or a simple json definition file can be used to create an interesting architecture. See json/netflixoss_arch.json (shown below) to see how to define an architecture without making code changes. The migration.go architecture is more complex as it steps through a sequence. If you figure out your own architecture in the form shown below it's going to be easy to carry forward as Spigo evolves. A big thanks is due to [Kurtis Kemple](https://github.com/kkemple) for cleaning up my initial javascript/D3 UI code and building simianviz as a single page app.

Keynote presentation at the O'Reilly Software Architecture Conference: Monitoring Microservices - A Challenge
http://www.slideshare.net/adriancockcroft/software-architecture-monitoring-microservices-a-challenge
Video of the 10 minute talk: https://youtu.be/smEuX-Hq6RI

```
{
    "arch": "netflixoss",
    "description":"A very simple Netflix service. See http://netflix.github.io/ to decode the package names",
    "version": "arch-0.0",
    "victim": "homepage",
    "services": [
        { "name": "cassSubscriber",   "package": "priamCassandra", "count": 6, "regions": 1, "dependencies": ["cassSubscriber", "eureka"]},
        { "name": "evcacheSubscriber","package": "store",          "count": 3, "regions": 1, "dependencies": []},
        { "name": "subscriber",       "package": "staash",         "count": 6, "regions": 1, "dependencies": ["cassSubscriber", "evcacheSubscriber"]},
        { "name": "login",            "package": "karyon",        "count": 18, "regions": 1, "dependencies": ["subscriber"]},
        { "name": "homepage",         "package": "karyon",        "count": 24, "regions": 1, "dependencies": ["subscriber"]},
        { "name": "wwwproxy",         "package": "zuul",           "count": 6, "regions": 1, "dependencies": ["login", "homepage"]},
        { "name": "www-elb",          "package": "elb",            "count": 0, "regions": 1, "dependencies": ["wwwproxy"]},
        { "name": "www",              "package": "denominator",    "count": 0, "regions": 0, "dependencies": ["www-elb"]}
    ]
}
```

For a single unscaled region, the above architecture is processed using spigo to produce json/netflixoss.json which is rendered using the single page app linked above or via a simpler local page local-d3-simianviz.html which can be used offline for quick tests with a local copy of d3:

![Netflixoss](png/netflixoss.png)

```
$ ./spigo -h
Usage of ./spigo:
  -a="netflixoss": Architecture to create or read, fsm, lamp, migration, netflixoss or json/????_arch.json
  -c=false: Collect metrics to json/<arch>_metrics.json and via http:
  -cpuprofile="": Write cpu profile to file
  -cpus=4:  Number of CPUs for Go runtime
  -d=10:    Simulation duration in seconds
  -f=false: Filter output names to simplify graph
  -g=false: Enable GraphML logging of nodes and edges to <arch>.graphml
  -j=false: Enable GraphJSON logging of nodes and edges to <arch>.json
  -m=false: Enable console logging of every message
  -p=100:   Pirate population for fsm or scale factor % for netflixoss etc.
  -r=false: Reload json/<arch>.json to setup architecture
  -s=0:     Stop creating microservices at this step, 0 = don't stop
  -u="1s":  Polling interval for Eureka name service
  -w=1:     Wide area regions

