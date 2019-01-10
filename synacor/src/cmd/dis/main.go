package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"instruction"
	"reader"
)

var (
	inputPath = flag.String("input", "", "input in binary format")
	lenFlag   = flag.Int("len", -1, "Number of shorts to print")
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

func dump(sr reader.Short) {
	for i := 0; *lenFlag == -1 || i < *lenFlag; i++ {
		addr := sr.Off()
		inst, bytesRead, err := instruction.Read(sr)
		if bytesRead == 0 {
			break
		} else if err != nil {
			fmt.Printf("%5d: error: %v\n", addr, err)
		} else {
			fmt.Printf("%5d: %s\n", addr, inst.String())
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

	dump(bio)
}
