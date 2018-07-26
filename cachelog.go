package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

func exactToBytes(str string) float32 {
	str = strings.TrimRight(str, "],")
	str = strings.TrimLeft(str, "[")
	n, err := strconv.ParseFloat(str, 32)
	if err != nil {
		log.Fatal(str, err)
	}
	return float32(n)
}

func toBytes(str string) float32 {
	n, err := strconv.ParseFloat(str[0:len(str)-2], 32)
	if err != nil {
		panic(err)
		log.Fatal(str, err)
	}
	m := str[len(str)-2:]
	switch {
	case m == "Kb":
		n = n * 1024
	case m == "Mb":
		n = n * 1024 * 1024
	case m == "Gb":
		n = n * 1024 * 1024 * 1024
	}
	return float32(n)
}

func deBracket(str string) string {
	if str[0] == '[' && str[len(str)-1] == ']' {
		str = str[1:]
		str = str[0 : len(str)-1]
	}
	return str
}

var firstTime time.Time
var lastTime time.Time

func checkContinuity(datefmt, date, hms string) {
	t, err := time.Parse(datefmt, date+" "+hms)
	if err != nil {
		log.Fatal(date, hms, err)
	}
	if lastTime == firstTime {
		lastTime = t
		return
	}
	if lastTime.Before(t) {
		lastTime = t
		return
	}
	if lastTime.Equal(t) {
		lastTime = t
		return
	}
	fmt.Fprintf(os.Stderr, "Backward hhmmss jump detected: %s %s\n", date, hms)
	lastTime = t
}

func main() {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Fprintf(os.Stderr, "PANIC %v\n%s", r, debug.Stack())
		}
		os.Exit(1)
	}()

	log.SetFlags(log.Lshortfile)
	ret := 0
	var heap, free, used float32
	pid := "none"

	flag.Usage = func() {
		man := []string{
			"NAME",
			"%s\n\nDESCRIPTION",
			"Reads object-server.log and outputs the memory usage metrics",
			"suitable for feeding into gnuplot.\n\nOPTIONS\n\n",
		}
		usage := fmt.Sprintf(strings.Join(man, "\n  "), os.Args[0])
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}
	exact := flag.Bool("exact", false, "(TODO) Get values from the the non SI-suffixed columns")
	flag.Parse()

	fmt.Println("pid date hhmmss heap free used")

	files := []string{"-"}
	if len(flag.Args()) > 0 {
		files = flag.Args()
	}

	var src io.Reader

	for _, file := range files {
		if file == "-" {
			src = os.Stdin
		} else {
			f, err := os.Open(file)
			if err != nil {
				log.Println(err)
				ret = 1
				continue
			}
			src = f
		}

		r := bufio.NewScanner(src)
		for r.Scan() {
			line := r.Text()
			if !strings.Contains(line, `space.cache-clean] Current heap size`) {
				continue
			}
			v3_9log := line[0] == '['
			tokens := strings.Split(line, " ")
			if pid != "none" && tokens[0] != pid {
				fmt.Println("")
			}
			pid = tokens[0]
			date, hhmmss := tokens[2], tokens[3]
			datefmt := "2006-01-02 15:04:05"
			if v3_9log {
				hhmmss = strings.TrimRight(hhmmss, ":")
				datefmt = "02/01/06 15:04:05"
			}
			checkContinuity(datefmt, date, hhmmss)
			if *exact {
				if v3_9log {
					// 0      1        2    3     4     56                     7       8    9     10    11           12    13     14         15    16    17           18    19     20
					// [1234] [broker] date hhmmss: (-!-)  0:[space.cache-clean] Current heap size: 410Mb [430186496], free: 8448Kb [8650752], used: 402Mb [421535744]. Water level: 500Mb
					heap = exactToBytes(tokens[11])
					free = exactToBytes(tokens[14])
					used = exactToBytes(tokens[17])
				} else {
					//  0     1      2          3        4   5 67                       8       9    10    11    12           13    14   15          16    17    18           19    20     21
					//  18002 broker 2016-01-19 11:25:44 UTC W  0:[-:space.cache-clean] Current heap size: 584Mb [613228544], free: 18Mb [19161088], used: 566Mb [594067456]. Water level: 700Mb
					heap = exactToBytes(tokens[12])
					free = exactToBytes(tokens[15])
					used = exactToBytes(tokens[18])
				}
			} else {
				if v3_9log {
					//fmt.Fprintf(os.Stderr, "line = '%s'\n", line)
					//fmt.Fprintf(os.Stderr, "tokens[9] = %s, tokens[12] = %s, tokens[15] = %s\n", tokens[9], tokens[12], tokens[15])
					// tokens[9] = size:, tokens[12] = free:, tokens[15] = used:
					heap = toBytes(tokens[10])
					free = toBytes(tokens[13])
					used = toBytes(tokens[16])
				} else {
					heap = toBytes(tokens[11])
					free = toBytes(tokens[14])
					used = toBytes(tokens[17])
				}
			}
			date_time := date + " " + hhmmss
			if v3_9log {
				t, err := time.Parse("02/01/06 15:04:05", date_time)
				if err != nil {
					panic(err)
				}
				date_time = t.Format("2006-01-02 15:04:05")
			}
			fmt.Println(deBracket(pid), date_time, heap, free, used)
		}
	}
	os.Exit(ret)
}

func stderr(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}
