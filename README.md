spigo and simianviz
===================

[![Join the chat at https://gitter.im/adrianco/spigo](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/adrianco/spigo?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

The name spigo is taken, but simianviz wasn't, so domains have been registered etc. and the name will transition over the coming months.

[![GoDoc](https://godoc.org/github.com/adrianco/spigo?status.svg)](https://godoc.org/github.com/adrianco/spigo) [Wiki Instructions](https://github.com/adrianco/spigo/wiki)

Simulate Protocol Interactions in Go using nanoservice actors - spigo

SIMulate Interactive Actor Network VIsualiZation - simianviz - also visualize the simian army in action (not yet implemented).

Current work in progress adds context to each message for request tracing, adds containers to the naming hierarchy and adds configurable filters to hide the extra levels when they aren't being used. The new -f option turns up filtering to produce a graph of services rather than nodes. Nodes also now have fake IP addresses. Docker compose yaml files can be converted to architecture json using
```
$ ./compose2arch -file compose.yaml > arch.json
```

Recent UI changes include pinning nodes so graphs can be stretched out and deleting nodes and edges by double-clicking on them.
[Run the netflixoss simulation in your browser](http://simianviz.divshot.io/netflixoss)

For a local installation of spigo, with no network dependencies, you can start the service and browse localhost:8000 using:
```
$ cd ui
$ npm install
$ npm run dev
```

Suitable for fairly large scale simulations, spigo runs well up to 100,000 independent nanoservice actors in a few GB of RAM. Three architectural models are implemented. One creates a peer to peer social network (fsm and pirates). Most others are based on a LAMP stack or NetflixOSS microservices in a more tree structured model loaded from an architecture definition file. The migration architecture is hard coded, starts with LAMP and ends with NetflixOSS.

Each nanoservice actor is a goroutine. to create 100,000 pirates, deliver 700,000 messages and wait to shut them all down again takes about 4 seconds. The resulting graph can be visualized via GraphML or rendered by saving to Graph JSON and viewing in a web browser via D3.

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

$ spigo -a netflixoss -d 5 -j
2015/05/25 12:16:12 Loading architecture from json_arch/netflixoss_arch.json
2015/05/25 12:16:12 netflixoss.edda: starting
2015/05/25 12:16:12 Architecture: netflixoss A very simple Netflix service. See http://netflix.github.io/ to decode the package names
2015/05/25 12:16:12 architecture: scaling to 100%
2015/05/25 12:16:12 Starting: {cassSubscriber priamCassandra 1 6 [cassSubscriber eureka]}
2015/05/25 12:16:12 netflixoss.us-east-1.zoneC.eureka.eureka.eureka2: starting
2015/05/25 12:16:12 netflixoss.us-east-1.zoneB.eureka.eureka.eureka1: starting
2015/05/25 12:16:12 netflixoss.us-east-1.zoneA.eureka.eureka.eureka0: starting
2015/05/25 12:16:12 Starting: {evcacheSubscriber store 1 3 []}
2015/05/25 12:16:12 Starting: {subscriber staash 1 6 [cassSubscriber evcacheSubscriber]}
2015/05/25 12:16:12 Starting: {login karyon 1 18 [subscriber]}
2015/05/25 12:16:12 Starting: {homepage karyon 1 24 [subscriber]}
2015/05/25 12:16:12 Starting: {wwwproxy zuul 1 6 [login homepage]}
2015/05/25 12:16:12 Starting: {www-elb elb 1 0 [wwwproxy]}
2015/05/25 12:16:12 Starting: {www denominator 0 0 [www-elb]}
2015/05/25 12:16:12 netflixoss.*.*.www.denominator.www0 activity rate  10ms
2015/05/25 12:16:14 chaosmonkey delete: netflixoss.us-east-1.zoneB.homepage.karyon.homepage4
2015/05/25 12:16:15 netflixoss.us-east-1.zoneB.eureka.eureka.eureka1:Forget netflixoss.us-east-1.zoneB.homepage.karyon.homepage4
2015/05/25 12:16:15 netflixoss.us-east-1.zoneB.eureka.eureka.eureka1:Forget netflixoss.us-east-1.zoneB.homepage.karyon.homepage4
2015/05/25 12:16:17 asgard: Shutdown
2015/05/25 12:16:17 netflixoss.us-east-1.zoneA.eureka.eureka.eureka0: closing
2015/05/25 12:16:17 netflixoss.us-east-1.zoneB.eureka.eureka.eureka1: closing
2015/05/25 12:16:17 netflixoss.us-east-1.zoneC.eureka.eureka.eureka2: closing
2015/05/25 12:16:17 spigo: complete
2015/05/25 12:16:17 netflixoss.edda: closing
```



Migration from LAMP to NetflixOSS
-----------
The orchestration to create this now uses a eureka discovery service per zone and has been heavily refactored.
[Run this simulation in your browser](http://simianviz.divshot.io/migration)

Start with a monolithic LAMP stack
![Migration ](png/migration-1-1.png)

Interpose Zuul proxy between load balancer and PHP monolith services
![Migration ](png/migration-2-1.png)

Replace single memcached with cross zone EVcache replicated memcached and change PHP to access MySQL via Staash (Storage Tier as a Service HTTP)
![Migration ](png/migration-3-1.png)
![Migration ](png/migration-3-2.png)

Add some Node based microservices between Zuul and Staash alongside PHP
![Migration ](png/migration-4-1.png)

Start a Cassandra cluster and connect to Staash alongside MySQL and evcache for data and access migration
![Migration ](png/migration-5-1.png)
![Migration ](png/migration-5-2.png)

Remove MySQL to be ready to go multi-region
![Migration ](png/migration-6-1.png)
![Migration ](png/migration-6-2.png)

Add a second region without connecting up cassandra
![Migration ](png/migration-7-1.png)
![Migration ](png/migration-7-2.png)
![Migration ](png/migration-7-3.png)

Connect regions together using multi-region Cassandra
![Migration ](png/migration-8-1.png)
![Migration ](png/migration-8-2.png)
![Migration ](png/migration-8-3.png)
![Migration ](png/migration-8-4.png)
![Migration ](png/migration-8-5.png)

Extend to six regions, an interesting visualization challenge
![Migration ](png/migration-9-1.png)
![Migration ](png/migration-9-2.png)
![Migration ](png/migration-9-3.png)
![Migration ](png/migration-9-5.png)

LAMP Stack Architecture
-----------
To create a starting point for architecture transitions, an AWS hosted LAMP stack is simulated. It has DNS feeding an ELB, then a horizontally scaled layer of PHP servers backed with a single memcached and a master slave pair of MySQL servers. The configuration is managed using a Eureka name service and logged by Edda. [Run this simulation in your browser](http://simianviz.divshot.io/lamp)

![LAMP stack](png/lamp.png)

Simple NetflixOSS Architecture and more complex Netflix Architecture
-----------
Simple simulations of the following AWS and NetflixOSS services are implemented. Edda collects the configuration and writes it to Json or Graphml. Eureka implements a service registry. Archaius contains global configuration data. Denominator simulates a global DNS endpoint. ELB generates traffic that is split across three availability zones. Zuul takes requests and routes it to the Karyon business logic layer. Karyon calls into the Staash data access layer, which calls PriamCassandra, which provides cross zone and cross region connections.

Each microservice is based on Karyon as the prototype to copy when creating a new microservice. The simulation passes get and put requests down the tree one at a time from Denominator. Get requests lookup the key in PriamCassandra and respond back up the tree. Put requests go down the tree only, and PriamCassandra replicates the put across all zones and regions.

There is a more complex architecture defined in json_arch/netflix_arch.json, which has two separate DNS endpoints for www and api, and three cassandra clusters. It provides a more realistic challenge for visualization.

[Run the netflixoss simulation in your browser](http://simianviz.divshot.io/netflixoss)

![Two Region NetflixOSS](png/netflixoss-w2-tooltip.png)

With the -m option all messages are logged as they are received. The time taken to deliver the message is shown
```
2015/03/01 13:16:09 netflixoss.us-east-1.ABC.api-elb.elb.api-elb0: gotocol: 18.9us Put remember me
2015/03/01 13:16:09 netflixoss.us-east-1.zoneC.apiproxy.zuul.apiproxy2: gotocol: 6.726us Put remember me
2015/03/01 13:16:09 netflixoss.us-east-1.zoneC.api.karyon.api23: gotocol: 6.002us Put remember me
2015/03/01 13:16:09 netflixoss.us-east-1.zoneC.turtle.staash.turtle2: gotocol: 5.891us Put remember me
2015/03/01 13:16:09 netflixoss.us-east-1.zoneC.cassTurtle.priamCassandra.cassTurtle11: gotocol: 5.798us Put remember me
2015/03/01 13:16:09 netflixoss.us-east-1.zoneA.cassTurtle.priamCassandra.cassTurtle0: gotocol: 8.393us Replicate remember me
2015/03/01 13:16:09 netflixoss.us-east-1.zoneB.cassTurtle.priamCassandra.cassTurtle1: gotocol: 30.158us Replicate remember me
2015/03/01 13:16:09 netflixoss.us-east-1.ABC.api-elb.elb.api-elb0: gotocol: 48.584us GetRequest why?
2015/03/01 13:16:09 netflixoss.us-east-1.zoneA.apiproxy.zuul.apiproxy3: gotocol: 13.474us GetRequest why?
2015/03/01 13:16:09 netflixoss.us-east-1.zoneA.api.karyon.api9: gotocol: 6.496us GetRequest why?
2015/03/01 13:16:09 netflixoss.us-east-1.zoneA.turtle.staash.turtle3: gotocol: 3.897us GetRequest why?
2015/03/01 13:16:09 netflixoss.us-east-1.zoneA.cassTurtle.priamCassandra.cassTurtle9: gotocol: 6.129us GetRequest why?
2015/03/01 13:16:09 netflixoss.us-east-1.zoneA.turtle.staash.turtle3: gotocol: 2.869us GetResponse because...
2015/03/01 13:16:09 netflixoss.us-east-1.zoneA.api.karyon.api9: gotocol: 2.169us GetResponse because...
2015/03/01 13:16:09 netflixoss.us-east-1.zoneA.apiproxy.zuul.apiproxy3: gotocol: 3.806us GetResponse because...
2015/03/01 13:16:09 netflixoss.us-east-1.ABC.api-elb.elb.api-elb0: gotocol: 2.272us GetResponse because...
2015/03/01 13:16:09 netflixoss.*.*.global-api-dns.denominator.global-api-dns0: gotocol: 2.422us GetResponse because...
```

100 Pirates
-----------
After seeding with two random friends GraphML rendered using yFiles
![100 pirates seeded with two random friends each](png/spigo100x2.png)

After chatting and making new friends rendered using graphJSON and D3
![100 pirates after chatting](png/spigo-100-json.png)

[Run this simulation in your browser](http://simianviz.divshot.io/fsm)

Spigo uses a common message protocol called Gotocol which contains a channel of the same type. This allows message listener endpoints to be passed around to dynamically create an arbitrary interconnection network.

Using terminology from Promise Theory each message also has an Imposition code that tells the receiver how to interpret it, and an Intention body string that can be used as a simple string, or to encode a more complex structured type or a Promise.

There is a central controller, the FSM (Flexible Simulation Manager or [Flying Spaghetti Monster](http://www.venganza.org/about/)), and a number of independent Pirates who listen to the FSM and to each other.

Current implementation creates the FSM and a default of 100 pirates, which can be set on the command line with -p=100. The FSM sends a Hello PirateNN message to name them which includes the FSM listener channel for back-chat. FSM then iterates through the pirates, telling each of them about two of their buddies at random to seed the network, giving them a random initial amount of gold coins, and telling them to start chatting to each other at a random pirate specific interval of between 0.1 and 10 seconds.

FSM can also reload from a json file that describes the nodes and edges in the network.

Either way FSM sleeps for a number of seconds then sends a Goodbye message to each. The Pirate responds to messages until it's told to chat, then it also wakes up at intervals and either tells one of its buddies about another one, or passes some of it's gold to a buddy until it gets a Goodbye message, then it quits and confirms by sending a Goodbye message back to the FSM. FSM counts down until all the Pirates have quit then exits.

The effect is that a complex randomized social graph is generated, with density increasing over time. This can then be used to experiment with trading, gossip and viral algorithms, and individual Pirates can make and break promises to introduce failure modes. Each pirate gets a random number of gold coins to start with, and can send them to buddies, and remember which benefactor buddy gave them how much.

Simulation is logged to a file spigo.graphml with the -g command line option or <arch>.json with the -j option. Inform messages are sent to a logger service from the pirates, which serializes writes to the file. The graphml format includes XML gibberish header followed by definitions of the node names and the edges that have formed between them. Graphml can be visualized using the yEd tool from yFiles. The graphJSON format is simpler and Javascript code to render it using D3 is in spigo.html.

There is a test program that exercises the Namedrop message, this is where the FSM or a Pirate passes on the name of a third party, and each Pirate builds up a buddy list of names and the listener channel for each buddy. Another test program tests the type conversions for JSON readings and writing.

The basic framework is in place, but more interesting behaviors, automonous running, and user input to control or stop the simulation haven't been added yet. [See the pdf for some Occam code](SkypeSim07.pdf) and results for the original version of this circa 2007.

Next steps include connecting the output directly to the browser over a websocket so the dynamic behavior of the graph can be seen in real time. A lot of refactoring has cleaned up the code and structure in preparation for more interesting features.

Jason Brown's list of interesting Gossip papers might contain something interesting to try and implement... http://softwarecarnival.blogspot.com/2014/07/gossip-papers.html

Benchmark result
================
At one point during setup FSM delivers five messages to each Pirate in turn, and the message delivery rate for that loop is measured at about 270,000 msg/sec. There are two additional shutdown messages per pirate in each run, plus whatever chatting occurs.
```
$ time spigo -d=0 -p=100000
2015/01/23 17:31:04 Spigo: population 100000 pirates
2015/01/23 17:31:05 fsm: Hello
2015/01/23 17:31:06 fsm: Talk amongst yourselves for 0
2015/01/23 17:31:07 fsm: Delivered 500000 messages in 1.865390635s
2015/01/23 17:31:07 fsm: Go away
2015/01/23 17:31:08 fsm: Exit
2015/01/23 17:31:08 spigo: fsm complete

real	0m3.968s
user	0m2.982s
sys	0m0.981s
```

Up to about 200,000 pirates time is linear with count. Beyond that it gradually slows down as my laptop runs out of memory.

