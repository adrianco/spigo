
## Scalability result
With go1.4.2 up to about 200,000 pirates time is linear with population. Beyond that it gradually slows down as my laptop runs out of memory.


## Message handling benchmark result

### Early code with go1.4 was fastest
At one point during setup FSM delivers five messages to each Pirate in turn, and the message delivery rate for that loop to generate, transfer and consume all the messages is measured at 500000/1.865=268,000 msg/sec. There are two additional shutdown messages per pirate in each run, plus whatever chatting occurs. 
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

### More complex and larger gotocol format added overhead
Later versions with request tracing additions to the message format and other changes slowed down to 198,000 msg/s with go1.4.2.
```
$ time spigo -d=0 -p=100000 -a fsm
2015/09/13 21:59:21 fsm: population 100000 pirates
2015/09/13 21:59:22 fsm: Talk amongst yourselves for 0
2015/09/13 21:59:24 fsm: Delivered 500000 messages in 2.522006488s
2015/09/13 21:59:24 fsm: Shutdown
2015/09/13 21:59:25 fsm: Exit
2015/09/13 21:59:25 spigo: complete

real	0m4.061s
user	0m6.790s
sys	0m1.678s
```

### Upgrading from go1.4.2 to go1.5.1 added overhead
After installing go1.5.1 and go build -a the message rate slowed down even more, to about 126,000 msg/s. To capture the difference cpu profiles were saved for each version, with no changes to the code base.

```
$ time ./spigo -d=0 -p=100000 -a fsm
2015/09/13 21:58:55 fsm: population 100000 pirates
2015/09/13 21:58:57 fsm: Talk amongst yourselves for 0
2015/09/13 21:59:01 fsm: Delivered 500000 messages in 3.957726898s
2015/09/13 21:59:01 fsm: Shutdown
2015/09/13 21:59:01 fsm: Exit
2015/09/13 21:59:01 spigo: complete

real	0m5.932s
user	0m7.471s
sys	0m1.728s
```

### Upgrading from go1.5.1 to go1.6.2
The codebase for fsm and pirates had not changed, although there are small changes to gotocol since the previous benchmarks were measured. Binaries for archived versions built for MacOS are saved along with CPU profiles for each version.

Running these tests back to back and repeating them on a Macbook Air running MacOS 10.11.3 El Capitan confirms the differences, however performance of the go1.5.1 binary appears to have improved slightly since the previous test.

```
$ time ./spigo.1.4.2 -a fsm -d 0 -p 100000
2016/04/21 20:03:27 fsm: population 100000 pirates
2016/04/21 20:03:28 fsm: Talk amongst yourselves for 0
2016/04/21 20:03:31 fsm: Delivered 500000 messages in 2.454943991s
2016/04/21 20:03:31 fsm: Shutdown
2016/04/21 20:03:31 fsm: Exit
2016/04/21 20:03:31 spigo: complete

real	0m3.950s
user	0m6.646s
sys	0m1.532s

$ time ./spigo.1.4.2 -a fsm -d 0 -p 100000
2016/04/21 20:03:34 fsm: population 100000 pirates
2016/04/21 20:03:35 fsm: Talk amongst yourselves for 0
2016/04/21 20:03:38 fsm: Delivered 500000 messages in 2.547419582s
2016/04/21 20:03:38 fsm: Shutdown
2016/04/21 20:03:38 fsm: Exit
2016/04/21 20:03:38 spigo: complete

real	0m4.017s
user	0m6.857s
sys	0m1.715s

$ time ./spigo.1.5.1 -a fsm -d 0 -p 100000
2016/04/21 20:03:44 fsm: population 100000 pirates
2016/04/21 20:03:45 fsm: Talk amongst yourselves for 0
2016/04/21 20:03:49 fsm: Delivered 500000 messages in 3.327230366s
2016/04/21 20:03:49 fsm: Shutdown
2016/04/21 20:03:49 fsm: Exit
2016/04/21 20:03:49 spigo: complete

real	0m4.756s
user	0m7.583s
sys	0m0.922s

$ time ./spigo.1.5.1 -a fsm -d 0 -p 100000
2016/04/21 20:03:51 fsm: population 100000 pirates
2016/04/21 20:03:52 fsm: Talk amongst yourselves for 0
2016/04/21 20:03:55 fsm: Delivered 500000 messages in 3.45077321s
2016/04/21 20:03:55 fsm: Shutdown
2016/04/21 20:03:56 fsm: Exit
2016/04/21 20:03:56 spigo: complete

real	0m4.861s
user	0m7.002s
sys	0m1.106s

$ time ./spigo.1.6.2 -a fsm -d 0 -p 100000
2016/04/21 20:04:00 fsm: population 100000 pirates
2016/04/21 20:04:01 fsm: Talk amongst yourselves for 0
2016/04/21 20:04:05 fsm: Delivered 500000 messages in 3.61466773s
2016/04/21 20:04:05 fsm: Shutdown
2016/04/21 20:04:05 fsm: Exit
2016/04/21 20:04:05 spigo: complete

real	0m5.314s
user	0m8.459s
sys	0m1.036s

$ time ./spigo.1.6.2 -a fsm -d 0 -p 100000
2016/04/21 20:04:21 fsm: population 100000 pirates
2016/04/21 20:04:23 fsm: Talk amongst yourselves for 0
2016/04/21 20:04:26 fsm: Delivered 500000 messages in 3.656322159s
2016/04/21 20:04:26 fsm: Shutdown
2016/04/21 20:04:26 fsm: Exit
2016/04/21 20:04:26 spigo: complete

real	0m4.883s
user	0m7.202s
sys	0m1.077s
```
### Analyzing CPU profiles

A brief look at the profile results didn't point out anything obvious, but I'm not sure how to interpret the differences.

```
$ cp ../spigo spigo.1.6.2
$ time ./spigo.1.6.2 -d=0 -p=100000 -a fsm -cpuprofile fsm.go162.profile
2016/04/20 23:55:50 fsm: population 100000 pirates
2016/04/20 23:55:51 fsm: Talk amongst yourselves for 0
2016/04/20 23:55:55 fsm: Delivered 500000 messages in 4.266286041s
2016/04/20 23:55:55 fsm: Shutdown
2016/04/20 23:55:56 fsm: Exit
2016/04/20 23:55:56 spigo: complete

real	0m5.569s
user	0m9.481s
sys	0m1.327s

$ cp ../spigo spigo.1.7.1
$ time ./spigo.1.7.1 -d=0 -p=100000 -a fsm -cpuprofile fsm.go171.profile
2016/10/17 20:14:29 fsm: population 100000 pirates
2016/10/17 20:14:30 fsm: Talk amongst yourselves for 0s
2016/10/17 20:14:35 fsm: Delivered 500000 messages in 5.162939826s
2016/10/17 20:14:35 fsm: Shutdown
2016/10/17 20:14:36 fsm: Exit
2016/10/17 20:14:36 spigo: complete

real	0m6.893s
user	0m9.721s
sys	0m2.762s

$ time ./spigo.1.7.1 -d=0 -p=100000 -a fsm -cpuprofile fsm.go171.profile
2016/10/17 20:14:49 fsm: population 100000 pirates
2016/10/17 20:14:50 fsm: Talk amongst yourselves for 0s
2016/10/17 20:14:54 fsm: Delivered 500000 messages in 4.043763355s
2016/10/17 20:14:54 fsm: Shutdown
2016/10/17 20:14:54 fsm: Exit
2016/10/17 20:14:54 spigo: complete

real	0m5.403s
user	0m7.860s
sys	0m2.638s

$ time ./spigo.1.7.1 -d=0 -p=100000 -a fsm -cpuprofile fsm.go171.profile
2016/10/17 20:15:20 fsm: population 100000 pirates
2016/10/17 20:15:22 fsm: Talk amongst yourselves for 0s
2016/10/17 20:15:26 fsm: Delivered 500000 messages in 4.154568125s
2016/10/17 20:15:26 fsm: Shutdown
2016/10/17 20:15:26 fsm: Exit
2016/10/17 20:15:26 spigo: complete

real	0m5.525s
user	0m8.732s
sys	0m2.243s

```
