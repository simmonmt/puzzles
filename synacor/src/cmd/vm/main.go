package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"instruction"
	"memory"
	"register"
	"symtab"
	"util"
)

var (
	ramPath         = flag.String("ram", "", "ram data load")
	symTabPath      = flag.String("symtab", "", "path to symbol table")
	verbose         = flag.Bool("verbose", false, "verbose")
	verboseFilePath = flag.String("verbose_file", "", "where to write verbose output; stdout if empty")
	dumpReg         = flag.Bool("dump_reg", false, "in verbose, dump registers after each instruction")
	initReg         = flag.String("init_reg", "", "initial register values, as r1=x,r2=y,...")
	startPCFlag     = flag.String("start_pc", "", "initial pc")
	haltPCFlag      = flag.String("halt_pc", "", "halt after executing this instruction")
	ramDumpPath     = flag.String("ram_dump", "", "where to dump ram on halt")
	overrideRAM     = flag.String("override_ram", "", "force some RAM values at start, as addr=val,addr=val,...")
)

func applyRAMOverrides(ram *memory.RAM, overrides string) error {
	for _, pair := range strings.Split(overrides, ",") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("bad pair %v", pair)
		}

		addr, err := strconv.ParseUint(parts[0], 10, 16)
		if err != nil {
			return fmt.Errorf("bad addr %v in pair %v", parts[0], pair)
		}

		val, err := strconv.ParseUint(parts[1], 10, 16)
		if err != nil {
			return fmt.Errorf("bad val %v in pair %v", parts[1], pair)
		}

		ram.Write(uint16(addr), uint16(val))
	}

	return nil
}

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

	if *overrideRAM != "" {
		if err := applyRAMOverrides(ram, *overrideRAM); err != nil {
			log.Fatalf("failed to apply RAM overrides: %v", err)
		}
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

	hupChan := make(chan os.Signal, 1)
	signal.Notify(hupChan, syscall.SIGHUP)

	var verboseWriter io.Writer
	if *verboseFilePath != "" {
		verboseFile, err := os.OpenFile(*verboseFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed to open verbose file %v: %v", *verboseFilePath, err)
		}
		defer verboseFile.Close()

		verboseWriter = verboseFile
	} else {
		verboseWriter = os.Stdout
	}

	stack := memory.NewStack()

	iCtx := instruction.Context{
		RAM:           ram,
		RegFile:       regFile,
		Stack:         stack,
		Verbose:       verbose,
		VerboseWriter: verboseWriter,
	}

	pc := startPC
	for {
		select {
		case <-hupChan:
			*verbose = !*verbose
			fmt.Printf("verbose now %v\n", *verbose)
		default:
		}

		reader := memory.NewRAMReader(ram, pc)

		inst, numRead, err := instruction.Read(reader)
		if err != nil {
			log.Fatalf("bad inst read at %v: %v", pc, err)
		}

		if *verbose {
			fmt.Fprintf(verboseWriter, "%30s: %s\n", util.AddrToName(pc, symTab),
				inst.ToString(symTab))
		}

		cb := instruction.CB{
			NPC: pc + uint16(numRead),
		}

		inst.Exec(&iCtx, &cb)

		if *verbose && *dumpReg {
			regFile.Dump(verboseWriter)
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
