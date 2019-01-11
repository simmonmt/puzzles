package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"instruction"
	"reader"
	"symtab"
)

var (
	inputPath  = flag.String("input", "", "input in binary format")
	lenFlag    = flag.Int("len", -1, "Number of shorts to print")
	symTabPath = flag.String("symtab", "", "path to symbol table")
	full       = flag.Bool("full", false, "include read bytes and raw addrs")
)

type BinIO struct {
	fp *os.File
}

func NewBinIO(path string) (*BinIO, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return &BinIO{fp}, nil
}

func (bio *BinIO) Read() (uint16, error) {
	b := [2]byte{}
	n, err := bio.fp.Read(b[:])
	if n != 2 {
		return 0, io.EOF
	} else if err != nil {
		return 0, err
	}

	var val uint16
	val |= uint16(b[0])
	val |= (uint16(b[1]) << 8)

	return val, nil
}

func (bio *BinIO) Off() uint16 {
	off, err := bio.fp.Seek(0, 1)
	if err != nil {
		panic("seek fail")
	}
	return uint16(off / 2)
}

func (bio *BinIO) Close() {
	if err := bio.fp.Close(); err != nil {
		panic(fmt.Sprintf("close fail: %v", err))
	}
}

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

func addrToName(addr uint16, st *symtab.SymTab) string {
	if st != nil {
		if ent, found := st.LookupAddr(addr); found {
			off := addr - ent.Start
			if off == 0 {
				return ent.Name
			}
			return fmt.Sprintf("%s+%d", ent.Name, off)
		}
	}
	return strconv.Itoa(int(addr))
}

func dump(sr reader.Short, st *symtab.SymTab) {
	saverReader := NewSaverReader(sr)

	for i := 0; *lenFlag == -1 || i < *lenFlag; i++ {
		addr := sr.Off()
		inst, numRead, instErr := instruction.Read(saverReader)
		if numRead == 0 {
			break
		}

		fmt.Printf("%30s:  ", addrToName(addr, st))

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
			fmt.Printf("error: %v\n", instErr)
		} else {
			fmt.Println(inst.ToString(st))
		}
	}
}

func main() {
	flag.Parse()

	if *inputPath == "" {
		log.Fatalf("--input is required")
	}

	bio, err := NewBinIO(*inputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer bio.Close()

	var st *symtab.SymTab
	if *symTabPath != "" {
		var err error
		if st, err = symtab.ReadFromPath(*symTabPath); err != nil {
			log.Fatal(err)
		}
	}

	dump(bio, st)
}
