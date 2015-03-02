#!/bin/sh

if [ -z "$1" ]; then
  echo "Usage: $0 log.txt"
  echo "log.txt is the file produced by cachelog"
  exit 1
fi
log=$1
gnuplot -persist -e "set xdata time;
set timefmt \"%d/%m/%y %H:%M:%S\";
set key autotitle columnhead;
set format x \"%H:%M\";
plot '$log' using 2:\"heap\" with lines, '' using 2:\"used\" with lines;"

#set terminal png size 3000,500;
