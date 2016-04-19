spigo and simianviz
===================

[![Join the chat at https://gitter.im/adrianco/spigo](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/adrianco/spigo?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

The name spigo is taken, but simianviz wasn't, so domains have been registered etc. and the name will transition over the coming months.

[![GoDoc](https://godoc.org/github.com/adrianco/spigo?status.svg)](https://godoc.org/github.com/adrianco/spigo) [Wiki Instructions](https://github.com/adrianco/spigo/wiki)

Simulate Protocol Interactions in Go using nanoservice actors - spigo

SIMulate Interactive Actor Network VIsualiZation - simianviz - also visualize the simian army in action.

[Run the netflixoss simulation in your browser](http://simianviz.surge.sh/netflixoss)

For a local installation of the above UI, with no network dependencies, you can start the service and browse localhost:8000 using:
```
$ cd ui
$ npm install
$ npm run dev
```

The [old README containing all the details](OLDREADME.md) is in
process of being cut up into multiple wiki pages and README files.

There were too many top level packages so a more hierachical directory
structure was setup.

```
top level
- spigo        # binary built for MacOS
- spigo.go     # main program
- actors       # code for packaged behaviors
- tooling      # support code
- ui           # visualization code using d3 and js
- misc         # scripts to run all tests and regenerate output
- json_arch    # architecture definition files
- json         # json dependency graph output
- json_metrics # flow, metrics and guesstimate output
- csv_metrics  # histograms saved as tables
- local-d3-simianviz.html # simple hackable d3 visualization
- png          # images for readme
- archived     # old files and packages
- gml          # old graphml dependency graphs
```

Docker compose version2 yaml files can be converted to architecture json using
```
$ cd compose2arch; go install

$ compose2arch -file myarch.yaml > json_arch/myarch.json
```

The basic framework is in place, but more interesting behaviors, automonous running, and user input to control or stop the simulation haven't been added yet. [See the pdf for some Occam code](misc/SkypeSim07.pdf) and results for the original version of this circa 2007.

Next steps include connecting the output directly to the browser over a websocket so the dynamic behavior of the graph can be seen in real time. A lot of refactoring has cleaned up the code and structure in preparation for more interesting features.

Jason Brown's list of interesting Gossip papers might contain something interesting to try and implement... http://softwarecarnival.blogspot.com/2014/07/gossip-papers.html

Keynote presentation at the O'Reilly Software Architecture Conference: Monitoring Microservices - A Challenge
http://www.slideshare.net/adriancockcroft/software-architecture-monitoring-microservices-a-challenge
Video of the 10 minute talk: https://youtu.be/smEuX-Hq6RI

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
  
