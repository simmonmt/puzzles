package main

import (
	"flag"
	"fmt"
	"log"

	"symtab"
	"util"
)

var (
	inputPath  = flag.String("input", "", "input in binary format")
	addrFlag   = flag.String("addr", "", "start addr")
	symTabPath = flag.String("symtab", "", "path to symbol table")
	dumpChar   = flag.Bool("dump_char", true, "if false, dump shorts rather than chars")
	indirect   = flag.Bool("indirect", false, "if true, dump from loc pointed to by addr")
)

func main() {
	flag.Parse()

	var symTab symtab.SymTab = &symtab.NoEntriesSymTab{}
	if *symTabPath != "" {
		var err error
		if symTab, err = symtab.ReadFromPath(*symTabPath); err != nil {
			log.Fatal(err)
		}
	}

	if *addrFlag == "" {
		log.Fatal("--addr is required")
	}

	addr, err := util.NameToAddr(*addrFlag, symTab)
	if err != nil {
		log.Fatalf("invalid --addr: %v", err)
	}

	bio, err := util.NewBinIO(*inputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer bio.Close()

	if err := bio.Seek(addr); err != nil {
		log.Fatal(err)
	}

	if *indirect {
		var err error
		addr, err = bio.Read()
		if err != nil {
			log.Fatal(err)
		}

		if err := bio.Seek(addr); err != nil {
			log.Fatal(err)
		}
	}

	objLen, err := bio.Read()
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < int(objLen); i++ {
		v, err := bio.Read()
		if err != nil {
			log.Fatal(err)
		}

		if *dumpChar {
			fmt.Printf("%s", string(byte(v&0xff)))
		} else {
			fmt.Printf("%5d  ", v)
		}
	}
	fmt.Println()
}
