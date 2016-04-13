grep args json/* | awk -F ':' '{print $3}' | tr -d '"[],' > misc/run.sh
ls */*_test.go | awk -F '/' '{print "cd " $1 ";go test;cd .."}' > misc/test.sh
