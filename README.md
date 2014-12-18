spigo
=====

Simulate Protocol Interactions in Go using nanoservice actors

Suitable for fairly large scale simulations, runs well up to 100,000 independent nanoservice actors (pirates). Each nanoservice is a goroutine, to create 100,000 pirates, deliver 600,000 messages and wait to shut them all down again takes about 2 seconds. The resulting graph can be visualized via GraphML or rendered be saving to Graph JSON and viewing in a web browser via D3.
```
$ spigo -h
Usage of spigo:
  -d=10: Simulation duration in seconds
  -g=false: Enable GraphML logging
  -j=false: Enable GraphJSON logging
  -p=100: Pirate population
  
$ spigo
Spigo population 100 pirates
Hello
Talk amongst yourselves for 10s
Delivered 400 messages in 610.781us
Go away
Pirate population: 0     
Exit
```
100 Pirates after seeding with two random friends GraphML rendered using yFiles
-----------
![100 pirates seeded with two random friends each](spigo100x2.png)

100 Pirates after chatting and making new friends rendered using graphJSON and D3
-----------
![100 pirates after chatting](spigo-100-json.png)

[Run spigo.html in your browser by clicking here](http://rawgit.com/adrianco/spigo/master/spigo.html)

Spigo uses a common message protocol called Gotocol which contains a channel of the same type. This allows message listener endpoints to be passed around to dynamically create an arbitrary interconnection network.

Using terminology from Promise Theory each message also has an Imposition code that tells the receiver how to interpret it, and an Intention body string that can be used as a simple string, or to encode a more complex structured type or a Promise.

There is a central controller, the FSM (Flexible Simulation Manager or [Flying Spaghetti Monster](http://www.venganza.org/about/)), and a number of independent Pirates who listen to the FSM and to each other.

Current implementation creates the FSM and a default of 100 pirates, which can be set on the command line with -p=100. The FSM sends a Hello PirateNN message to name them which includes the FSM listener channel for back-chat. FSM then iterates through the pirates, telling each of them about two of their buddies at random to seed the network, and telling them to start Chatting to each other at a random pirate specific interval of between 1 and 60 seconds. FSM sleeps for a number of seconds then sends a Goodbye message to each. The Pirate responds to messages until it's told to Chat, then it also wakes up every second and tells one of its buddies about another one until it gets a Goodbye message, then it quits and confirms by sending a Goodbye message back to the FSM. FSM counts down until all the Pirates have quit then exits.

The effect is that a complex randomized social graph is generated, with density increasing over time. This can then be used to experiment with trading, gossip and viral algorithms, and individual Pirates can make and break promises to introduce failure modes. Each pirate gets a random number of gold coins to start with, and can send them to buddies, and remember which benefactor buddy gave them how much.

Simulation is logged to a file spigo.graphml with the -g command line option or spigo.json with the -j option. Inform messages are sent back to the FSM from the pirates, and the FSM serializes writes to the file. The graphml format includes XML gibberish header followed by definitions of the node names and the edges that have formed between them. Graphml can be visualized using the yEd tool from yFiles. The graphJSON format is simpler and Javascript code to render it using D3 is in spigo.html.

There is a test program that exercises the Namedrop message, this is where the FSM or a Pirate passes on the name of a third party, and each Pirate builds up a buddy list of names and the listener channel for each buddy.

The basic framework is in place, but the interesting behaviors, automonous running, and user input to control or stop the simulation haven't been added yet. [See the pdf for some Occam code](SkypeSim07.pdf) and results for the original version of this circa 2007.

Next steps include reading in a saved graphJSON to form the initial graph and connecting the output directly to the browser over a websocket so the dynamic behavior of the graph can be seen in real time.

Jason Brown's list of interesting Gossip papers might contain something interesting to try and implement... http://softwarecarnival.blogspot.com/2014/07/gossip-papers.html

Benchmark result
================
At one point during setup FSM delivers three messages to each Pirate in turn, and the message delivery rate for that loop is measured at about 650,000 msg/sec. There are three additional messages per pirate in each run, plus whatever chatting occurs.
```
$ time spigo -d=0 -p=100000
Spigo population 100000 pirates
Hello
Talk amongst yourselves for 0
Delivered 300000 messages in 465.231966ms
Go away
Pirate population: 0        
Exit

real	0m2.122s
user	0m1.574s
sys	0m0.466s
```

Up to about 200,000 pirates time is linear with count. Beyond that it slows down and with 1,000,000 initialization takes about 47s, and the process uses about 5GB RAM so there's probably an inefficiency in the way the map of names and channels is being created, or its taking a long time to steal 5GB RAM from other things on my 8GB RAM MacBook Air.

