## json format dependency graph files


Typical run including -j to write json graph files here

```
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

Using the step option, multiple runs can create different outputs in a sequence that can be stepped through.
The architecture for Migration is hard wired by the migration package for each step.

[Run this simulation in your browser](http://simianviz.surge.sh/migration)

Start with a monolithic LAMP stack
![Migration ](../png/migration-1-1.png)

Interpose Zuul proxy between load balancer and PHP monolith services
![Migration ](../png/migration-2-1.png)

Replace single memcached with cross zone EVcache replicated memcached and change PHP to access MySQL via Staash (Storage Tier as a Service HTTP)
![Migration ](../png/migration-3-1.png)
![Migration ](../png/migration-3-2.png)

Add some Node based microservices between Zuul and Staash alongside PHP
![Migration ](../png/migration-4-1.png)

Start a Cassandra cluster and connect to Staash alongside MySQL and evcache for data and access migration
![Migration ](../png/migration-5-1.png)
![Migration ](../png/migration-5-2.png)

Remove MySQL to be ready to go multi-region
![Migration ](../png/migration-6-1.png)
![Migration ](../png/migration-6-2.png)

Add a second region without connecting up cassandra
![Migration ](../png/migration-7-1.png)
![Migration ](../png/migration-7-2.png)
![Migration ](../png/migration-7-3.png)

Connect regions together using multi-region Cassandra
![Migration ](../png/migration-8-1.png)
![Migration ](../png/migration-8-2.png)
![Migration ](../png/migration-8-3.png)
![Migration ](../png/migration-8-4.png)
![Migration ](../png/migration-8-5.png)

Extend to six regions, an interesting visualization challenge
![Migration ](../png/migration-9-1.png)
![Migration ](../png/migration-9-2.png)
![Migration ](../png/migration-9-3.png)
![Migration ](../png/migration-9-5.png)

LAMP Stack Architecture
-----------
To create a starting point for architecture transitions, an AWS hosted LAMP stack is simulated. It has DNS feeding an ELB, then a horizontally scaled layer of PHP servers backed with a single memcached and a master slave pair of MySQL servers. The configuration is managed using a Eureka name service and logged by Edda. [Run this simulation in your browser](http://simianviz.surge.sh/lamp)

![LAMP stack](../png/lamp.png)

Simple NetflixOSS Architecture and more complex Netflix Architecture
-----------
Simple simulations of the following AWS and NetflixOSS services are implemented. Edda collects the configuration and writes it to Json or Graphml. Eureka implements a service registry. Archaius contains global configuration data. Denominator simulates a global DNS endpoint. ELB generates traffic that is split across three availability zones. Zuul takes requests and routes it to the Karyon business logic layer. Karyon calls into the Staash data access layer, which calls PriamCassandra, which provides cross zone and cross region connections.

Each microservice is based on Karyon as the prototype to copy when creating a new microservice. The simulation passes get and put requests down the tree one at a time from Denominator. Get requests lookup the key in PriamCassandra and respond back up the tree. Put requests go down the tree only, and PriamCassandra replicates the put across all zones and regions.

There is a more complex architecture defined in json_arch/netflix_arch.json, which has two separate DNS endpoints for www and api, and three cassandra clusters. It provides a more realistic challenge for visualization.

[Run the netflixoss simulation in your browser](http://simianviz.surge.sh/netflixoss)

![Two Region NetflixOSS](../png/netflixoss-w2-tooltip.png)

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
