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

func main() {
	ret := 0
	re := regexp.MustCompile(`\[space\.cache-clean\] Current heap size`)
	var heap, free, used float32
	pid := "none"

	exact := flag.Bool("exact", false, "Get values from the the non SI-suffixed columns")
	flag.Parse()

	fmt.Println("pid date time heap free used")

	files := []string{"-"}
	if len(os.Args) >= 2 {
		files = os.Args[1:]
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
