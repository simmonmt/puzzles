package main

import (
	"fmt"
)

var (
	coins = map[int]string{
		2: "red",
		7: "concave",
		3: "corroded",
		9: "blue",
		5: "shiny",
	}
)

type Permuter struct {
	values []int
	used   []bool // necessary?
	idxs   []int
}

func NewPermuter(values []int) *Permuter {
	p := &Permuter{
		values: values,
		used:   make([]bool, len(values)),
		idxs:   make([]int, len(values)),
	}
	p.Reset()
	return p
}

func (p *Permuter) Reset() {
	for i := range p.idxs {
		p.idxs[i] = i
		p.used[i] = true
	}
}

func (p *Permuter) Values() []int {
	out := make([]int, len(p.values))
	for i := range p.idxs {
		out[i] = p.values[p.idxs[i]]
	}
	return out
}

func (p *Permuter) Advance() bool {
	// Find the latest index that can advance ignoring its children
	var cand int
	for cand = len(p.values) - 1; cand >= 0; cand-- {
		// can cand advance?
		canAdvance := false
		for i := p.idxs[cand] + 1; i < len(p.values); i++ {
			if !p.used[i] {
				canAdvance = true
				break
			}
		}
		if canAdvance {
			break
		}

		// cand cannot advance so clear it
		p.used[p.idxs[cand]] = false
		p.idxs[cand] = -1
	}

	if cand < 0 {
		return false
	}

	for i := cand; i < len(p.values); i++ {
		var target int
		for target = p.idxs[i] + 1; target < len(p.values); target++ {
			if !p.used[target] {
				break
			}
		}
		if target >= len(p.values) {
			panic("no target")
		}

		if p.idxs[i] >= 0 {
			p.used[p.idxs[i]] = false
		}
		p.idxs[i] = target
		p.used[target] = true
	}

	return true
}

func verify(v []int) bool {
	return v[0]+(v[1]*v[2]*v[2])+v[3]*v[3]*v[3]-v[4] == 399
}

func main() {
	values := []int{}
	for k := range coins {
		values = append(values, k)
	}

	permuter := NewPermuter(values)

	for {
		values := permuter.Values()
		if verify(values) {
			for _, value := range values {
				fmt.Println(coins[value])
			}
		}

		if !permuter.Advance() {
			break
		}
	}
}
