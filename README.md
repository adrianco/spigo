spigo and simianviz
===================

[![Join the chat at https://gitter.im/adrianco/spigo](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/adrianco/spigo?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

The name spigo is taken, but simianviz wasn't, so domains have been registered etc. and the name will transition over the coming months.

[![GoDoc](https://godoc.org/github.com/adrianco/spigo?status.svg)](https://godoc.org/github.com/adrianco/spigo) [Wiki Instructions](https://github.com/adrianco/spigo/wiki)

Simulate Protocol Interactions in Go using nanoservice actors - spigo

SIMulate Interactive Actor Network VIsualiZation - simianviz - also visualize the simian army in action (not yet implemented).

Docker compose version2 yaml files can be converted to architecture json using
```
$ ./compose2arch -file compose.yaml > arch.json
```

[Run the netflixoss simulation in your browser](http://simianviz.surge.sh/netflixoss)

For a local installation of the above UI, with no network dependencies, you can start the service and browse localhost:8000 using:
```
$ cd ui
$ npm install
$ npm run dev
```

The [old README containing all the details](OLDREADME.md) is in
process of being cut up into multiple wiki pages and README files.

There are too many top level packages so a more hierachical directory
structure is being setup.
