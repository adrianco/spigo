
Suitable for fairly large scale simulations, spigo runs well up to 100,000 independent nanoservice actors in a few GB of RAM. Three types of architectural models are implemented. One creates a peer to peer social network (fsm and pirates). Most others are based on a LAMP stack or NetflixOSS microservices in a more tree structured model loaded from an architecture definition file. The migration architecture is hard coded, starts with LAMP and ends with NetflixOSS.

Each nanoservice actor is a goroutine. to create 100,000 pirates, deliver 700,000 messages and wait to shut them all down again took about 4 seconds. The resulting graph can be visualized via GraphML or rendered by saving to Graph JSON and viewing in a web browser via D3.

A few lines of code or a simple json definition file can be used to create an interesting architecture. See json/netflixoss_arch.json (shown below) to see how to define an architecture without making code changes. The migration.go architecture is more complex as it steps through a sequence. If you figure out your own architecture in the form shown below it's going to be easy to carry forward as Spigo evolves. A big thanks is due to [Kurtis Kemple](https://github.com/kkemple) for cleaning up my initial javascript/D3 UI code and building simianviz as a single page app.


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

![Netflixoss](../png/netflixoss.png)
