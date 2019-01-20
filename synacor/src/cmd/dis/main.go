package main

import (
	"flag"
	"fmt"
	"log"

	"comment"
	"instruction"
	"reader"
	"symtab"
	"util"
)

var (
	inputPath    = flag.String("input", "", "input in binary format")
	lenFlag      = flag.Int("len", -1, "Number of shorts to print")
	commentsPath = flag.String("comments", "", "path to comment registry")
	symTabPath   = flag.String("symtab", "", "path to symbol table")
	full         = flag.Bool("full", false, "include read bytes and raw addrs")
)

type SaverReader struct {
	r  reader.Short
	vr [5]uint16
	n  int
}

func NewSaverReader(r reader.Short) *SaverReader {
	return &SaverReader{
		r: r,
	}
}

func (sr *SaverReader) Read() (uint16, error) {
	v, err := sr.r.Read()
	if err != nil {
		return 0, err
	}

	if sr.n < len(sr.vr) {
		sr.vr[sr.n] = v
		sr.n++
	}

	return v, err
}

func (sr *SaverReader) Off() uint16 {
	return sr.r.Off()
}

// Returns a slice containing the last values read. Invalidated upon next Read
// call.
func (sr *SaverReader) ValuesRead() []uint16 {
	values := sr.vr[0:sr.n]
	sr.n = 0
	return values
}

func dump(sr reader.Short, st symtab.SymTab, cReg comment.Registry) {
	saverReader := NewSaverReader(sr)

	var curBlock *comment.Comment

	for i := 0; *lenFlag == -1 || i < *lenFlag; i++ {
		addr := sr.Off()

		if block := cReg.GetBlock(int(addr)); block != nil {
			if curBlock != nil {
				panic("nested block")
			}
			curBlock = block

			padLen := 33
			if *full {
				padLen += 33 // 8 + 4*6 + 1
			}

			fmt.Println()
			for _, line := range block.Lines {
				fmt.Printf("%*s// %s\n", padLen, "", line)
			}
		}

		inst, numRead, instErr := instruction.Read(saverReader)
		if numRead == 0 {
			break
		}

		fmt.Printf("%30s:  ", util.AddrToName(addr, st))

		if *full {
			fmt.Printf("%5d:  ", addr)

			vr := saverReader.ValuesRead()
			if len(vr) > 4 {
				panic("long read")
			}
			for _, v := range vr {
				fmt.Printf("%5d ", v)
			}
			for i := len(vr); i < 4; i++ {
				fmt.Print("      ")
			}

			fmt.Print(" ")
		}

		if instErr != nil {
			fmt.Println("error: ", instErr)
		} else {
			fmt.Printf("%-30s", inst.ToString(st))
			if comment, found := cReg.GetSingle(int(addr)); found {
				fmt.Printf("// %s", comment)
			}
			fmt.Println()
		}

		if curBlock != nil && int(addr) == curBlock.End {
			fmt.Println()
			curBlock = nil
		}
	}
}

func main() {
	flag.Parse()

	if *inputPath == "" {
		log.Fatalf("--input is required")
	}

	bio, err := util.NewBinIO(*inputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer bio.Close()

	var symTab symtab.SymTab = &symtab.NoEntriesSymTab{}
	if *symTabPath != "" {
		var err error
		if symTab, err = symtab.ReadFromPath(*symTabPath); err != nil {
			log.Fatal(err)
		}
	}

	var commentRegistry comment.Registry = &comment.NullRegistry{}
	if *commentsPath != "" {
		var err error
		if commentRegistry, err = comment.ReadFromPath(*commentsPath, symTab); err != nil {
			log.Fatalf("failed to read comment registry: %v", err)
		}
	}

	dump(bio, symTab, commentRegistry)
}
