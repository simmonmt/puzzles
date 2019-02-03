package main

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func posLess(a, b Pos) bool {
	if a.X < b.X {
		return true
	} else if a.X > b.X {
		return false
	}
	return a.Y < b.Y
}

type Orbs []Orb

func (a Orbs) Len() int      { return len(a) }
func (a Orbs) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Orbs) Less(i, j int) bool {
	if posLess(a[i].Pos, a[j].Pos) {
		return true
	} else if posLess(a[j].Pos, a[i].Pos) {
		return false
	}

	return a[i].Val < a[j].Val
}

func TestNeighbors(t *testing.T) {
	tests := []struct {
		start     Orb
		neighbors []Orb
	}{
		{Orb{Pos{3, 0}, 100}, nil},
		{Orb{Pos{0, 1}, 100},
			[]Orb{
				Orb{Pos: Pos{X: 1, Y: 0}, Val: 108},
				Orb{Pos: Pos{X: 1, Y: 0}, Val: 800},
				Orb{Pos: Pos{X: 1, Y: 2}, Val: 104},
				Orb{Pos: Pos{X: 1, Y: 2}, Val: 400},
				Orb{Pos: Pos{X: 2, Y: 1}, Val: 1100},
			}},
		{Orb{Pos{1, 2}, 1},
			[]Orb{
				Orb{Pos: Pos{X: 0, Y: 1}, Val: 4},
				Orb{Pos: Pos{X: 0, Y: 1}, Val: 5},
				Orb{Pos: Pos{X: 1, Y: 0}, Val: 8},
				Orb{Pos: Pos{X: 2, Y: 1}, Val: 11},
			}},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d: %+v", i, test.start),
			func(t *testing.T) {
				neighbors := board.Neighbors(test.start)
				sort.Sort(Orbs(neighbors))

				if !reflect.DeepEqual(neighbors, test.neighbors) {
					t.Errorf("board.Neighbors(%+v) = %+v, want %+v",
						test.start, neighbors, test.neighbors)
				}
			})
	}
}
