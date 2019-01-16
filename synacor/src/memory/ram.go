package memory

import (
	"fmt"
	"io/ioutil"
)

type RAM [32768]uint16

func (r *RAM) Read(addr uint16) uint16 {
	if (addr & 0x8000) != 0 {
		panic("big addr")
	}

	return r[addr]
}

func (r *RAM) Write(addr, val uint16) {
	if (addr & 0x8000) != 0 {
		panic("big addr")
	}

	r[addr] = val
}

func DumpRAM(ram *RAM, path string) error {
	vals := make([]byte, len(ram)*2)
	for i, s := range ram {
		vals[i*2] = byte(s & 0xff)
		vals[i*2+1] = byte((s >> 8) & 0xff)
	}

	return ioutil.WriteFile(path, vals, 0644)
}

type RAMReader struct {
	ram *RAM
	off uint16
}

func NewRAMReader(ram *RAM, off uint16) *RAMReader {
	return &RAMReader{
		ram: ram,
		off: off,
	}
}

func (r *RAMReader) Read() (uint16, error) {
	if int(r.off) >= len(r.ram) {
		return 0, fmt.Errorf("out of range")
	}
	val := r.ram.Read(r.off)
	r.off++
	return val, nil
}

func (r *RAMReader) Off() uint16 {
	return r.off
}
