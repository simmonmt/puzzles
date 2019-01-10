package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"instruction"
	"memory"
	"register"
)

var (
	ramPath = flag.String("ram", "", "ram data load")
	verbose = flag.Bool("verbose", false, "verbose")
)

func initRAM(path string) (*memory.RAM, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	ram := &memory.RAM{}

	var b [2]byte
	for off := 0; ; off++ {
		n, _ := fp.Read(b[:])
		if n != 2 {
			break
		}

		var val uint16
		val |= uint16(b[0])
		val |= (uint16(b[1]) << 8)

		ram.Write(uint16(off), val)
	}

	return ram, nil
}

func main() {
	flag.Parse()

	if *ramPath == "" {
		log.Fatalf("--ram is required")
	}

	ram, err := initRAM(*ramPath)
	if err != nil {
		log.Fatal(err)
	}

	regFile := &register.File{}
	stack := memory.NewStack()

	iCtx := instruction.Context{
		RAM:     ram,
		RegFile: regFile,
		Stack:   stack,
		Verbose: *verbose,
	}

	var pc uint16
	for {
		reader := memory.NewRAMReader(ram, pc)

		inst, numRead, err := instruction.Read(reader)
		if err != nil {
			log.Fatalf("bad inst read at %v: %v", pc, err)
		}

		if *verbose {
			fmt.Printf("%5d: %s\n", pc, inst)
		}

		cb := instruction.CB{
			NPC: pc + uint16(numRead),
		}

		inst.Exec(&iCtx, &cb)
		if cb.Hlt {
			fmt.Printf("hlt requested at %v\n", pc)
			break
		}

		pc = cb.NPC
	}
}
