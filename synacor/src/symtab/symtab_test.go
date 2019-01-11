package symtab

import (
	"reflect"
	"testing"
)

func TestAdd(t *testing.T) {
	st := New()
	if err := st.Add(2, 5, "A"); err != nil {
		t.Errorf(`st.Add(2, 5, "A") = %v, want nil`, err)
		return
	}
	if err := st.Add(9, 18, "A"); err == nil {
		t.Errorf(`st.Add(9, 18, "A") = %v, want non-nil`, err)
		return
	}
}

func TestLookupAddr(t *testing.T) {
	st := New()
	st.Add(2, 5, "A")
	st.Add(8, 10, "B")

	tests := []struct {
		addr  uint16
		ent   SymEnt
		found bool
	}{
		{addr: 1, found: false},
		{addr: 2, ent: SymEnt{Name: "A", Start: 2, End: 5}, found: true},
		{addr: 3, ent: SymEnt{Name: "A", Start: 2, End: 5}, found: true},
		{addr: 5, ent: SymEnt{Name: "A", Start: 2, End: 5}, found: true},
		{addr: 6, found: false},
		{addr: 7, found: false},
		{addr: 8, ent: SymEnt{Name: "B", Start: 8, End: 10}, found: true},
		{addr: 9, ent: SymEnt{Name: "B", Start: 8, End: 10}, found: true},
		{addr: 10, ent: SymEnt{Name: "B", Start: 8, End: 10}, found: true},
		{addr: 11, found: false},
	}

	for _, test := range tests {
		ent, found := st.LookupAddr(test.addr)
		if !found {
			if test.found {
				t.Errorf("LookupAddr(%v) = !found, want %+v, %v",
					test.addr, test.ent, test.found)
			}
		} else { // found
			if !test.found {
				t.Errorf("LookupAddr(%v) = %+v, %v, want !found",
					test.addr, test.ent, found)
			} else if !reflect.DeepEqual(ent, test.ent) {
				t.Errorf("LookupAddr(%v) = %+v, %v, want %+v, %v",
					test.addr, ent, found, test.ent, test.found)
			}
		}
	}
}
