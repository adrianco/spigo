./spigo -a aws_ac_ra_web -d 3 -j -p 100
./spigo -d 10 -j -a cassandra -c
./spigo -a cassandra -d 2 -j -s 1
./spigo -a cassandra -d 2 -j -s 2 -p 200
./spigo -a cassandra -d 2 -j -s 3 -p 400
./spigo -a cassandra -d 4 -j -s 4 -p 100 -w 2
./spigo -a cassandra -d 4 -j -s 5 -p 100 -w 3
./spigo -a cassandra -d 4 -j -s 6 -p 100 -w 4
./spigo -a cassandra -d 4 -j -s 7 -p 100 -w 5
./spigo -a cassandra -d 4 -j -s 8 -p 100 -w 6
./spigo -a cassandra -d 4 -j -s 9 -p 200 -w 6
./spigo -a composeV2 -d 2 -j
./spigo -a composeV2 -d 2 -j -p 200 -s 1
./spigo -a composeV2 -d 2 -j -p 100 -s 2 -w 2
./spigo -a container -j -f
./spigo -a container -j -d 4 -s 1
./spigo -a container -j -d 4 -s 2 -p 200
./spigo -a container -j -d 4 -s 3 -p 150 -w 2
./spigo -a fsm -d 10 -j -w 2
./spigo -a fsm -d 30 -j -p 100 -s 1
./spigo -a fsm -d 10 -j -p 200 -s 2
./spigo -a fsm -d 10 -j -p 300 -s 3
./spigo -a fsm -d 10 -j -p 400 -s 4
./spigo -a fsm -d 10 -j -p 500 -s 5
./spigo -a lamp -d 2 -j
./spigo -a lamp -d 1 -j -s 1
./spigo -a lamp -d 5 -j -p 200 -s 2
./spigo -a lamp -d 5 -j -p 300 -s 3
./spigo -a lamp -d 5 -j -p 300 -s 4 -w 2
./spigo -a lamp -d 5 -j -p 300 -s 5 -w 3
./spigo -a lamp -d 5 -j -p 200 -s 6 -w 4
./spigo -a lamp -d 5 -j -p 200 -s 7 -w 5
./spigo -a lamp -d 5 -j -p 200 -s 8 -w 6
./spigo -a lamp -d 5 -j -p 100 -s 9 -w 6 -f
./spigo -a migration -d 5 -j
./spigo -a migration -d 5 -j -s 1
./spigo -a migration -d 3 -j -s 2
./spigo -a migration -d 3 -j -s 3
./spigo -a migration -d 3 -j -s 4
./spigo -a migration -d 3 -j -s 5
./spigo -a migration -d 3 -j -s 6
./spigo -a migration -d 3 -j -s 7 -w 2
./spigo -a migration -d 3 -j -s 8 -w 2
./spigo -a migration -d 3 -j -s 9 -w 3
./spigo -a netflix -d 5 -j -c
./spigo -a netflix -d 3 -j -p 200 -s 1
./spigo -a netflix -d 3 -j -p 300 -s 2
./spigo -a netflix -d 3 -j -p 100 -s 3 -w 2
./spigo -a netflix -d 3 -j -p 100 -s 4 -w 3
./spigo -a netflix -d 10 -j -p 300 -s 5 -w 6
./spigo -a netflix -d 3 -j -p 100 -s 6 -w 6
./spigo -a netflix -d 3 -j -p 200 -s 7 -w 6 -f
./spigo -a netflix -d 10 -j -p 300 -s 8 -w 6
./spigo -a netflix -d 60 -j -p 400 -s 9 -w 6 -u=10s
./spigo -d 2 -j -w 1 -a netflixoss -c -p 100
./spigo -a netflixoss -d 5 -j -p 200 -s 1
./spigo -a netflixoss -d 5 -j -p 100 -s 2 -w 2
./spigo -a netflixoss -d 5 -j -p 100 -s 3 -w 3
./spigo -a netflixoss -d 5 -j -p 100 -s 4 -w 4
./spigo -a netflixoss -d 5 -j -p 100 -s 5 -w 5
./spigo -a netflixoss -d 5 -j -p 100 -s 6 -w 6
./spigo -a netflixoss -d 5 -j -p 200 -s 7 -w 6
./spigo -a netflixoss -d 5 -j -p 300 -s 8 -w 6
./spigo -a netflixoss -d 5 -j -p 400 -s 9 -w 6 -f=true
./spigo -a riak -j -d 2
./spigo -d 2 -a riak -j -s 1 -w 2 -c
./spigo -d 2 -a riak -j -s 2 -w 2 -p 200
./spigo -a simpleV2 -d 2 -j
./spigo -a storage -j -c -d 60
./spigo -a storage -d 2 -p 200 -s 1 -j
./spigo -a testyaml -d 2 -j
./spigo -a yogi -d 2 -j -c -f
./spigo -a yogi -d 5 -j -p 100 -s 1 -w 1
./spigo -a yogi -d 5 -j -p 100 -s 2 -w 2
./spigo -a yogi -d 5 -j -p 100 -s 3 -w 3
