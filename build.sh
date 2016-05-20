grep args json/* | awk -F ':' '{print $3}' | tr -d '"[],' > run.sh
ls */*_test.go | awk -F '/' '{print "cd " $1 ";go test;cd .."}' >test.sh
