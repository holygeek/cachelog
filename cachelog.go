package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func toBytes(str string) float32 {
	n, err := strconv.ParseFloat(str[0:len(str)-2], 32)
	if err != nil {
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

func checkContinuity(date, hms string) {
	t, err := time.Parse("02/01/06 15:04:05", date+" "+hms)
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
	fmt.Fprintf(os.Stderr, "Backward time jump detected: %s %s\n", date, hms)
	lastTime = t
}

func main() {
	ret := 0
	re := regexp.MustCompile(`\[space\.cache-clean\] Current heap size`)
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

	fmt.Println("pid date time heap free used")

	files := []string{"-"}
	if len(flags.Args()) > 0 {
		files = flags.Args()
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
			if !re.MatchString(line) {
				continue
			}
			tokens := strings.Split(line, " ")
			if pid != "none" && tokens[0] != pid {
				fmt.Println("")
			}
			pid = tokens[0]
			date, time := tokens[2], tokens[3][0:len(tokens[3])-1]
			checkContinuity(date, time)
			if *exact {
				fmt.Println("TODO use exact")
			} else {
				heap = toBytes(tokens[10])
				free = toBytes(tokens[13])
				used = toBytes(tokens[16])
			}
			fmt.Println(deBracket(pid), date, time, heap, free, used)
		}
	}
	os.Exit(ret)
}
