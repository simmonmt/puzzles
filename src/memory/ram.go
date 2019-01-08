package memory

import "fmt"

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
