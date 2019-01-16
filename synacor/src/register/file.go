package register

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type File struct {
	reg [8]uint16
}

func (f *File) Get(num uint) uint16 {
	if num >= uint(len(f.reg)) {
		return 0
	}
	return f.reg[num]
}

func (f *File) Set(num uint, val uint16) {
	if num >= uint(len(f.reg)) {
		return
	}
	f.reg[num] = val
}

func (f *File) Dump() {
	for i := 0; i < len(f.reg); i++ {
		if i != 0 {
			fmt.Print(", ")
		}
		fmt.Printf("r%d=%d", i, f.reg[i])
	}
	fmt.Println()
}

var (
	regPattern = regexp.MustCompile(`^r(\d+)=(\d+)$`)
)

func InitFromSpec(specs string) (*File, error) {
	rf := &File{}

	for _, spec := range strings.Split(specs, ",") {
		parts := regPattern.FindStringSubmatch(spec)
		if parts == nil {
			return nil, fmt.Errorf("failed to parse spec %v", spec)
		}

		regNum, err := strconv.ParseUint(parts[1], 10, 8)
		if err != nil {
			return nil, fmt.Errorf("failed to parse reg num in spec %v: %v",
				spec, err)
		}

		if regNum > 7 {
			return nil, fmt.Errorf("illegal reg num %v in spec %v", regNum, spec)
		}

		val, err := strconv.ParseUint(parts[2], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to parse reg val in spec %v: %v",
				spec, err)
		}

		rf.Set(uint(regNum), uint16(val))
	}

	return rf, nil
}
