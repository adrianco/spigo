grep args json/* | awk -F ':' '{print $3}' | tr -d '"[],' > run.sh
