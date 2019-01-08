package main

import (
	"flag"
	"fmt"
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
		return 0, fmt.Errorf("short read")
	} else if err != nil {
		return 0, err
	}

	var val uint16
	val |= uint16(b[0])
	val |= (uint16(b[1]) << 8)

	return val, nil
}

func (bio *BinIO) Close() {
	if err := bio.fp.Close(); err != nil {
		panic(fmt.Sprintf("close fail: %v", err))
	}
}

func dump(sr reader.Short) {
	addr := 0
	for *lenFlag == -1 || addr < *lenFlag {
		inst, argLen, err := instruction.Read(sr)
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}

		fmt.Printf("%5d: %s\n", addr, inst.String())
		addr += argLen
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
