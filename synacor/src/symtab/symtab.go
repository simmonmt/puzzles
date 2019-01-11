package symtab

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/HuKeping/rbtree"
)

type SymEnt struct {
	Start, End uint16
	Name       string
}

func (e *SymEnt) Less(than rbtree.Item) bool {
	return e.Start < than.(*SymEnt).Start
}

type SymTab struct {
	tree   *rbtree.Rbtree
	byName map[string]*SymEnt
}

func New() *SymTab {
	return &SymTab{
		tree:   rbtree.New(),
		byName: map[string]*SymEnt{},
	}
}

var (
	symtabPattern = regexp.MustCompile(`^(\w+)\s+([0-9]+)(?:-([0-9]))?(?:\s*#.*)$`)
)

func Read(r io.Reader) (*SymTab, error) {
	st := New()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := symtabPattern.FindStringSubmatch(line)
		if parts == nil {
			panic("nomatch")
		} else {
			panic("match")
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return st, nil
}

func (st *SymTab) Add(start, end uint16, name string) error {
	if _, found := st.byName[name]; found {
		return fmt.Errorf("%v already exists in table", name)
	}

	e := &SymEnt{start, end, name}
	st.byName[name] = e
	st.tree.Insert(e)

	return nil
}

func (st *SymTab) LookupAddr(addr uint16) (SymEnt, bool) {
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
