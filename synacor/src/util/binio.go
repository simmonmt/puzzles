package util

import (
	"fmt"
	"io"
	"os"
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

func (bio *BinIO) Seek(off uint16) error {
	_, err := bio.fp.Seek(int64(off)*2, 0)
	return err
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
