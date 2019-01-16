package util

import (
	"fmt"
	"strconv"

	"patterns"
	"symtab"
)

func AddrToName(addr uint16, st symtab.SymTab) string {
	if ent, found := st.LookupAddr(uint(addr)); found {
		off := addr - uint16(ent.Start)
		if off == 0 {
			return ent.Name
		}
		return fmt.Sprintf("%s+%d", ent.Name, off)
	}
	return strconv.Itoa(int(addr))
}

func NameToAddr(name string, st symtab.SymTab) (uint16, error) {
	if addr, err := strconv.ParseUint(name, 0, 16); err == nil {
		return uint16(addr), nil
	}

	parts := patterns.NameWithOptionalOffsetPattern.FindStringSubmatch(name)
	if parts == nil {
		return 0, fmt.Errorf("match failure")
	}

	symName, offStr := parts[1], parts[2]

	symEnt, found := st.LookupName(symName)
	if !found {
		return 0, fmt.Errorf("unknown symbol %v", symName)
	}

	if offStr == "" {
		return uint16(symEnt.Start), nil
	}

	off, err := strconv.ParseUint(offStr, 0, 16)
	if err != nil {
		return 0, fmt.Errorf("bad offset %v", offStr)
	}

	if symEnt.Start+uint(off) > symEnt.End {
		return 0, fmt.Errorf("offset in %v extends beyond %v", name, symName)
	}

	return uint16(symEnt.Start) + uint16(off), nil
}
