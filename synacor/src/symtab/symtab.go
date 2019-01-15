package symtab

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/HuKeping/rbtree"
)

type SymEnt struct {
	Start, End uint
	Name       string
}

func (e *SymEnt) Less(than rbtree.Item) bool {
	return e.Start < than.(*SymEnt).Start
}

func (e *SymEnt) OffStr(addr uint) string {
	if addr == e.Start {
		return e.Name
	}
	return fmt.Sprintf("%s+%d", e.Name, addr-e.Start)
}

type SymTab interface {
	Add(name string, start, end uint) error
	LookupAddr(addr uint) (SymEnt, bool)
	LookupName(name string) (SymEnt, bool)
}

type NoEntriesSymTab struct{}

func (s *NoEntriesSymTab) Add(name string, start, end uint) error { return fmt.Errorf("unsupported") }
func (s *NoEntriesSymTab) LookupAddr(addr uint) (SymEnt, bool)    { return SymEnt{}, false }
func (s *NoEntriesSymTab) LookupName(name string) (SymEnt, bool)  { return SymEnt{}, false }

type symTabImpl struct {
	tree   *rbtree.Rbtree
	byName map[string]*SymEnt
}

func New() SymTab {
	return &symTabImpl{
		tree:   rbtree.New(),
		byName: map[string]*SymEnt{},
	}
}

var (
	symtabPattern = regexp.MustCompile(`^(\w+)\s+([0-9]+)(?:-([0-9]+))?(?:\s*#.*)?$`)
)

func Read(r io.Reader) (SymTab, error) {
	st := New()

	scanner := bufio.NewScanner(r)
	for lineNum := 1; scanner.Scan(); lineNum++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := symtabPattern.FindStringSubmatch(line)
		if parts == nil {
			return nil, fmt.Errorf("%d: parse fail", lineNum)
		}

		name := parts[1]
		startStr := parts[2]
		endStr := parts[3]
		if endStr == "" {
			endStr = startStr
		}

		start, err := strconv.ParseUint(startStr, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("%d: parse start fail: %v", lineNum, err)
		}
		end, err := strconv.ParseUint(endStr, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("%d: parse end fail: %v", lineNum, err)
		}

		st.Add(name, uint(start), uint(end))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return st, nil
}

func ReadFromPath(path string) (SymTab, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	return Read(fp)
}

func (st *symTabImpl) Add(name string, start, end uint) error {
	if _, found := st.byName[name]; found {
		return fmt.Errorf("%v already exists in table", name)
	}

	e := &SymEnt{start, end, name}
	st.byName[name] = e
	st.tree.Insert(e)

	return nil
}

func (st *symTabImpl) LookupAddr(addr uint) (SymEnt, bool) {
	pivot := &SymEnt{Start: addr}

	var match *SymEnt
	st.tree.Descend(pivot, func(i rbtree.Item) bool {
		e := i.(*SymEnt)
		if addr <= e.End {
			match = e
		}
		return false
	})

	if match == nil {
		return SymEnt{}, false
	}
	return *match, true
}

func (st *symTabImpl) LookupName(name string) (SymEnt, bool) {
	ent, found := st.byName[name]
	if !found {
		return SymEnt{}, false
	}

	return *ent, true
}
