package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	dictPath = flag.String("dict", "/usr/share/dict/american-english", "dict path")
)

func findInputWords(input string) map[string]bool {
	words := map[string]bool{}

	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		words[strings.ToLower(scanner.Text())] = true
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return words
}

func dumpOutput(w io.Writer, input string, goodWords map[string]bool) {
	lineScanner := bufio.NewScanner(strings.NewReader(input))
	for lineScanner.Scan() {
		line := lineScanner.Text()

		start := true
		wordScanner := bufio.NewScanner(strings.NewReader(line))
		wordScanner.Split(bufio.ScanWords)
		for wordScanner.Scan() {
			word := wordScanner.Text()
			if _, found := goodWords[strings.ToLower(word)]; found {
				if start {
					start = false
				} else {
					fmt.Fprint(w, " ")
				}
				fmt.Fprint(w, word)
			}
		}
		fmt.Fprintln(w)
	}
}

func main() {
	flag.Parse()
	args := flag.Args()

	inputArr, err := ioutil.ReadFile(args[0])
	if err != nil {
		log.Fatal(err)
	}
	input := string(inputArr)

	inputWords := findInputWords(input)

	goodInputWords := map[string]bool{}

	{
		dictFile, err := os.Open(*dictPath)
		if err != nil {
			log.Fatal(err)
		}
		defer dictFile.Close()

		scanner := bufio.NewScanner(dictFile)
		for scanner.Scan() {
			word := strings.ToLower(scanner.Text())
			if _, found := inputWords[word]; found {
				goodInputWords[word] = true
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatalf("read failed: %v", err)
		}
	}

	dumpOutput(os.Stdout, input, goodInputWords)
}
