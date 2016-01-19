#!/bin/sh
set -e
me=$(basename $0)

usage() {
	echo "Usage: $me [-h] [-t <black|white>] [-W width] [-H height] <log.txt>

log.txt is the file produced by cachelog

OPTIONS
    -t <black|white>
	Use black or white theme. Default is white.
    -h
	Show this help message
    -d <duration>
	Limit xrange max to <duration>, e.g: -d '1 hour'
    -s <since>
	Limit xrange min to <since>, e.g: -s '1 day ago'
    -T <title prefix>

    -t <[white]|black>

    -W <width>

    -H <height>"
}

titlePrefix=""
width=1000
height=500
theme=white
since=
duration=
while getopts d:hW:H:s:T:t: opt
do
	case "$opt" in
		d)
			duration=$OPTARG
			;;
		W)
			width=$OPTARG
			;;
		H)
			height=$OPTARG
			;;
		s)
			since=$OPTARG
			;;
		T)
			titlePrefix="$OPTARG "
			;;
		t)
			theme="$OPTARG"
			;;
		h)
			usage
			exit
			;;
		\?)
			echo "$me: Unknown option '$opt'"
			exit 1
			;;
	esac
done
shift $(($OPTIND -1))

if [ -z "$1" ]; then
  usage
  exit 1
fi

if [ "$theme" = "black" ]; then
  fg=white
  bg=black
elif [ "$theme" = "white" ]; then
  bg=white
  fg=black
else
  echo "Unknown theme: $theme"
  exit 1
fi


log=$1
#png_file=$log-$(date -Iseconds).png
png_file=$log.png
xmin=
xmax=
timefmt='%d/%m/%y %H:%M:%S'
log_start=
if [ -n "$since" ]; then
	xmin=\"$(date --utc -d "$since" +"$timefmt")\"
	if [ -n "$duration" ]; then
		xmax=\"$(date --utc -d "$since + $duration" +"$timefmt")\"
	fi
	log_start=$(date --utc -d "$since" +"$timefmt")
else
  log_start=$(head -2 "$log"|tail -1|awk '{print $2" "$3}')
fi

if [ -n "$xmin" -o -n "$xmax" ]; then
	xrange='set xrange ['$xmin':'$xmax']'
fi

title="${titlePrefix}Log start $log_start"
gnuplot -persist -e "set xdata time;
set timefmt \"$timefmt\";
set title '$title' textcolor rgb \"$fg\";
set key autotitle columnhead textcolor rgb \"$fg\";
set format x \"%l%p\n%a\";
$xrange;
set xlabel \"Time (UTC)\" textcolor rgb \"$fg\";
set ylabel \"Memory use in bytes\" textcolor rgb \"$fg\";
set terminal png size $width,$height background rgb\"$bg\";
set border lw 1 lc rgb \"$fg\";
set xtics textcolor rgb \"$fg\";
set ytics textcolor rgb \"$fg\";
set grid linecolor rgb \"gray\";

plot '$log' using 2:\"heap\" with lines, '' using 2:\"used\" with lines;" >$png_file
echo qiv $png_file &&
qiv $png_file

#https://gist.github.com/tetsuok/2639931
## Change colors of elements in Gnuplot
#
## change a color of border.
#set border lw 3 lc rgb "white"
#
## change text colors of  tics
#set xtics textcolor rgb "white"
#set ytics textcolor rgb "white"
#
## change text colors of labels
#set xlabel "X" textcolor rgb "white"
#set ylabel "Y" textcolor rgb "white"
#
## change a text color of key
#set key textcolor rgb "white"
