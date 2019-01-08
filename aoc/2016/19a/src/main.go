package main

import (
	"flag"
	"fmt"
	"log"

	"elf"
	"logger"
)

var (
	numElves = flag.Int("num_elves", -1, "number of elves")
	verbose  = flag.Bool("verbose", false, "verbose")
)

func main() {
	flag.Parse()
	logger.Init(*verbose)

	if *numElves == -1 {
		log.Fatal("--num_elves is required")
	}

	elves := elf.InitElves(*numElves)
	name := elf.Play(elves)
	fmt.Println(name)
}
