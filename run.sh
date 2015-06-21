./spigo -a aws_ac_ra_web -d 3 -j -p 100
./spigo -a fsm -d 10 -j -p 100
./spigo -a fsm -d 30 -j -p 100 -s 1
./spigo -a fsm -d 10 -j -p 200 -s 2
./spigo -a fsm -d 10 -j -p 300 -s 3
./spigo -a fsm -d 10 -j -p 400 -s 4
./spigo -a fsm -d 10 -j -p 500 -s 5
./spigo -a lamp -d 5 -j -c
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
./spigo -d 5 -j
./spigo -a netflixoss -d 5 -j -p 200 -s 1
./spigo -a netflixoss -d 5 -j -p 100 -s 2 -w 2
./spigo -a netflixoss -d 5 -j -p 100 -s 3 -w 3
./spigo -a netflixoss -d 5 -j -p 100 -s 4 -w 4
./spigo -a netflixoss -d 5 -j -p 100 -s 5 -w 5
./spigo -a netflixoss -d 5 -j -p 100 -s 6 -w 6
./spigo -a netflixoss -d 5 -j -p 200 -s 7 -w 6
./spigo -a netflixoss -d 5 -j -p 300 -s 8 -w 6
./spigo -a netflixoss -d 5 -j -p 400 -s 9 -w 6
