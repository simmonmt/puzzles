package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"

	"instruction"
	"memory"
	"register"
	"symtab"
	"util"
)

var (
	ramPath     = flag.String("ram", "", "ram data load")
	symTabPath  = flag.String("symtab", "", "path to symbol table")
	verbose     = flag.Bool("verbose", false, "verbose")
	dumpReg     = flag.Bool("dump_reg", false, "in verbose, dump registers after each instruction")
	initReg     = flag.String("init_reg", "", "initial register values, as r1=x,r2=y,...")
	startPCFlag = flag.String("start_pc", "", "initial pc")
	haltPCFlag  = flag.String("halt_pc", "", "halt after executing this instruction")
	ramDumpPath = flag.String("ram_dump", "", "where to dump ram on halt")
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

func parsePCFlagsOrDie(symTab symtab.SymTab) (startPC, haltPC uint16) {
	if *startPCFlag != "" {
		var err error
		if startPC, err = util.NameToAddr(*startPCFlag, symTab); err != nil {
			log.Fatalf("invalid --start_pc value: %v", err)
		}
	}

	haltPC = math.MaxUint16
	if *haltPCFlag != "" {
		var err error
		if haltPC, err = util.NameToAddr(*haltPCFlag, symTab); err != nil {
			log.Fatalf("invalid --start_pc value: %v", err)
		}
	}

	return
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

	var symTab symtab.SymTab = &symtab.NoEntriesSymTab{}
	if *symTabPath != "" {
		var err error
		if symTab, err = symtab.ReadFromPath(*symTabPath); err != nil {
			log.Fatal(err)
		}
	}

	regFile := &register.File{}
	if *initReg != "" {
		var err error
		regFile, err = register.InitFromSpec(*initReg)
		if err != nil {
			log.Fatalf("failed to init registers: %v", err)
		}
	}

	startPC, haltPC := parsePCFlagsOrDie(symTab)

	stack := memory.NewStack()

	iCtx := instruction.Context{
		RAM:     ram,
		RegFile: regFile,
		Stack:   stack,
		Verbose: *verbose,
	}

	pc := startPC
	for {
		reader := memory.NewRAMReader(ram, pc)

		inst, numRead, err := instruction.Read(reader)
		if err != nil {
			log.Fatalf("bad inst read at %v: %v", pc, err)
		}

		if *verbose {
			fmt.Printf("%30s: %s\n", util.AddrToName(pc, symTab),
				inst.ToString(symTab))
		}

		cb := instruction.CB{
			NPC: pc + uint16(numRead),
		}

		inst.Exec(&iCtx, &cb)

		if *dumpReg {
			regFile.Dump()
		}

		if haltPC != math.MaxUint16 && haltPC == pc {
			fmt.Printf("hlt requested by flag at %v\n", pc)
			break
		}

		if cb.Hlt {
			fmt.Printf("hlt requested at %v\n", pc)
			break
		}

		pc = cb.NPC
	}

	if *ramDumpPath != "" {
		log.Printf("dumping RAM to %v", *ramDumpPath)
		if err := memory.DumpRAM(ram, *ramDumpPath); err != nil {
			log.Fatalf("failed to dump RAM: %v", err)
		}
	}
}
