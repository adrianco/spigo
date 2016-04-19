
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

Later versions with request tracing additions to the message format and other changes slowed down to about 200,000 msg/s.
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

After installing go1.5.1 and go build -a the message rate slowed down even more, to about 125,000 msg/s
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

Up to about 200,000 pirates time is linear with count. Beyond that it gradually slows down as my laptop runs out of memory.
