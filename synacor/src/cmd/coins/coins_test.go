package main

import (
	"reflect"
	"strconv"
	"testing"
)

func intsToStr(a []int) string {
	str := ""
	for _, v := range a {
		if str != "" {
			str += ","
		}
		str += strconv.Itoa(v)
	}
	return str
}

func TestPermuter(t *testing.T) {
	p := NewPermuter([]int{2, 7, 3, 9, 5})

	expecteds := [][]int{
		[]int{2, 7, 3, 9, 5},
		[]int{2, 7, 3, 5, 9},
		[]int{2, 7, 9, 3, 5},
		[]int{2, 7, 9, 5, 3},
		[]int{2, 7, 5, 3, 9},
		[]int{2, 7, 5, 9, 3},
	}

	for i, expected := range expecteds {
		got := p.Values()
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("%d advances, got %v, wanted %v", i, got, expected)
		}

		if !p.Advance() {
			t.Errorf("%d advance = false, want true", i)
		}
	}

	p.Reset()
	numPerms := 1
	seen := map[string]bool{}
	for ; ; numPerms++ {
		key := intsToStr(p.Values())
		if _, found := seen[key]; found {
			t.Errorf("repeat permutation %v", key)
			break
		}
		seen[key] = true

		if !p.Advance() {
			break
		}
	}

	if numPerms != 120 {
		t.Errorf("expected 120 permutations, got %v", numPerms)
	}
}
