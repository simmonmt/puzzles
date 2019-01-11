package symtab

import (
	"reflect"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	in := `
		# comment
		a 4
		b 8-10  # another comment`

	st, err := Read(strings.NewReader(in))
	if err != nil {
		t.Errorf("read failed: %v", err)
	}

	tests := []struct {
		name  string
		found bool
		ent   SymEnt
	}{
		{"a", true, SymEnt{Name: "a", Start: 4, End: 4}},
		{"b", true, SymEnt{Name: "b", Start: 8, End: 10}},
	}

	for _, test := range tests {
		ent, found := st.LookupName(test.name)
		if found != test.found || (test.found && !reflect.DeepEqual(ent, test.ent)) {
			t.Errorf(`LookupName("%v") = %v, %v, want %v, %v`,
				test.name, ent, found, test.ent, test.found)
		}
	}
}

func TestAdd(t *testing.T) {
	st := New()
	if err := st.Add("A", 2, 5); err != nil {
		t.Errorf(`st.Add("A", 2, 5) = %v, want nil`, err)
		return
	}
	if err := st.Add("A", 9, 18); err == nil {
		t.Errorf(`st.Add("A", 9, 18) = %v, want non-nil`, err)
		return
	}
}

func TestLookupAddr(t *testing.T) {
	st := New()
	st.Add("a", 2, 5)
	st.Add("b", 8, 10)

	tests := []struct {
		addr  uint16
		found bool
		ent   SymEnt
	}{
		{addr: 1, found: false},
		{addr: 2, found: true, ent: SymEnt{Name: "a", Start: 2, End: 5}},
		{addr: 3, found: true, ent: SymEnt{Name: "a", Start: 2, End: 5}},
		{addr: 5, found: true, ent: SymEnt{Name: "a", Start: 2, End: 5}},
		{addr: 6, found: false},
		{addr: 7, found: false},
		{addr: 8, found: true, ent: SymEnt{Name: "b", Start: 8, End: 10}},
		{addr: 9, found: true, ent: SymEnt{Name: "b", Start: 8, End: 10}},
		{addr: 10, found: true, ent: SymEnt{Name: "b", Start: 8, End: 10}},
		{addr: 11, found: false},
	}

	for _, test := range tests {
		ent, found := st.LookupAddr(test.addr)
		if found != test.found || (test.found && !reflect.DeepEqual(ent, test.ent)) {
			t.Errorf("LookupAddr(%v) = %+v, %v, want %+v, %v",
				test.addr, ent, found, test.ent, test.found)
		}
	}
}

func TestLookupName(t *testing.T) {
	st := New()
	st.Add("a", 2, 5)
	st.Add("b", 8, 10)

	tests := []struct {
		name  string
		found bool
		ent   SymEnt
	}{
		{name: "a", found: true, ent: SymEnt{Name: "a", Start: 2, End: 5}},
		{name: "b", found: true, ent: SymEnt{Name: "b", Start: 8, End: 10}},
		{name: "c", found: false},
	}

	for _, test := range tests {
		ent, found := st.LookupName(test.name)
		if found != test.found || (test.found && !reflect.DeepEqual(ent, test.ent)) {
			t.Errorf(`LookupName("%v") = %+v, %v, want %+v, %v`,
				test.name, ent, found, test.ent, test.found)
		}
	}
}
