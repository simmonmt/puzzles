package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"astar"
)

type CellType int

const (
	CELL_END CellType = iota
	CELL_START
	CELL_PLUS
	CELL_MINUS
	CELL_STAR
	CELL_VALUE
)

func (t CellType) String() string {
	switch t {
	case CELL_END:
		return "end"
	case CELL_START:
		return "start"
	case CELL_PLUS:
		return "plus"
	case CELL_MINUS:
		return "minus"
	case CELL_STAR:
		return "star"
	case CELL_VALUE:
		return "value"
	default:
		panic("unknown")
	}
}

func (t CellType) IsOp() bool {
	return t == CELL_PLUS || t == CELL_MINUS || t == CELL_STAR
}

func (t CellType) HasValue() bool {
	return t == CELL_VALUE || t == CELL_END
}

type Cell struct {
	Type CellType
	Val  int
}

type Pos struct {
	X, Y int
}

type Orb struct {
	Pos Pos
	Val int
}

func OrbFromString(str string) Orb {
	parts := strings.Split(str, ",")

	x, err := strconv.Atoi(parts[0])
	if err != nil {
		panic("bad x")
	}
	y, err := strconv.Atoi(parts[1])
	if err != nil {
		panic("bad y")
	}
	val, err := strconv.Atoi(parts[2])
	if err != nil {
		panic("bad val")
	}

	return Orb{Pos{x, y}, val}
}

func (o *Orb) ToString() string {
	return fmt.Sprintf("%d,%d,%d", o.Pos.X, o.Pos.Y, o.Val)
}

type Board struct {
	cells [][]Cell
}

func (b *Board) get(pos Pos) *Cell {
	if pos.X < 0 || pos.Y < 0 || pos.X >= len(b.cells[0]) || pos.Y >= len(b.cells) {
		return nil
	}

	return &b.cells[pos.Y][pos.X]
}

var (
	deltas = []Pos{Pos{-1, 0}, Pos{1, 0}, Pos{0, -1}, Pos{0, 1}}
)

func (b *Board) allNeighbors(pos Pos) []Pos {
	out := []Pos{}
	for _, d := range deltas {
		cand := Pos{X: pos.X + d.X, Y: pos.Y + d.Y}
		if cell := b.get(cand); cell != nil {
			out = append(out, cand)
		}
	}
	return out
}

func makeOrb(start Orb, pos Pos, op CellType, val int) Orb {
	out := Orb{
		Pos: pos,
		Val: start.Val,
	}

	switch op {
	case CELL_PLUS:
		out.Val += val
	case CELL_MINUS:
		out.Val -= val
	case CELL_STAR:
		out.Val *= val
	default:
		panic(fmt.Sprintf("unknown: %s", op))
	}

	return out
}

func (b *Board) Neighbors(start Orb) []Orb {
	if c := b.get(start.Pos); c.Type == CELL_END {
		return nil
	}

	out := []Orb{}
	for _, nPos := range b.allNeighbors(start.Pos) {
		nCell := b.get(nPos)
		if !nCell.Type.IsOp() {
			continue
		}

		for _, vPos := range b.allNeighbors(nPos) {
			if vPos == start.Pos {
				continue
			}

			vCell := b.get(vPos)
			if !vCell.Type.HasValue() {
				continue
			}

			cand := makeOrb(start, vPos, nCell.Type, vCell.Val)
			if cand.Val <= 0 {
				continue
			}

			out = append(out, cand)
		}
	}
	return out
}

var (
	board = &Board{
		cells: [][]Cell{
			// *  8  - D1
			// 4  * 11  *
			// +  4  - 18
			// A  -  9  *
			[]Cell{Cell{CELL_STAR, 0}, Cell{CELL_VALUE, 8}, Cell{CELL_MINUS, 0}, Cell{CELL_END, 1}},
			[]Cell{Cell{CELL_VALUE, 4}, Cell{CELL_STAR, 0}, Cell{CELL_VALUE, 11}, Cell{CELL_STAR, 0}},
			[]Cell{Cell{CELL_PLUS, 0}, Cell{CELL_VALUE, 4}, Cell{CELL_MINUS, 0}, Cell{CELL_VALUE, 18}},
			[]Cell{Cell{CELL_START, 0}, Cell{CELL_MINUS, 0}, Cell{CELL_VALUE, 9}, Cell{CELL_STAR, 0}},
		},
	}
)

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func dist(x1, y1, x2, y2 int) int {
	return abs(x1-x2) + abs(y1-y2)
}

type astarHelper struct{}

func (h *astarHelper) AllNeighbors(start string) []string {
	so := OrbFromString(start)

	out := []string{}
	for _, neighbor := range board.Neighbors(so) {
		out = append(out, neighbor.ToString())
	}
	return out
}

func (h *astarHelper) EstimateDistance(start, end string) uint {
	so := OrbFromString(start)
	eo := OrbFromString(end)

	return uint(dist(so.Pos.X, so.Pos.Y, eo.Pos.X, eo.Pos.Y))
}

func (h *astarHelper) NeighborDistance(n1, n2 string) uint {
	return 1
}

func (h *astarHelper) GoalReached(cand, goal string) bool {
	return cand == goal
}

func advance(from, to Pos) []Pos {
	if from.X == to.X {
		if from.Y > to.Y {
			return []Pos{Pos{from.X, from.Y - 1}}
		} else {
			return []Pos{Pos{from.X, from.Y + 1}}
		}
	} else if from.Y == to.Y {
		if from.X > to.X {
			return []Pos{Pos{from.X - 1, from.Y}}
		} else {
			return []Pos{Pos{from.X + 1, from.Y}}
		}
	}

	return []Pos{
		Pos{from.X, to.Y},
		Pos{to.X, from.Y},
	}
}

func checkOpPos(start Orb, cand Pos, end Orb) bool {
	typ := board.get(cand).Type
	if !typ.IsOp() {
		//fmt.Printf("rejecting cand %+v; not op: %+v\n", cand, board.get(cand))
		return false
	}

	//fmt.Printf("cand %+v\n", cand)
	o := makeOrb(start, end.Pos, typ, board.get(end.Pos).Val)
	//fmt.Printf("start %+v cand %+v end %+v o %+v\n", start, cand, end, o)
	return end.Val == o.Val
}

func findOpPos(start, end Orb) Pos {
	for _, cand := range advance(start.Pos, end.Pos) {
		if checkOpPos(start, cand, end) {
			return cand
		}
	}
	panic("no route")
}

func main() {
	start := Orb{Pos{0, 3}, 22}
	goal := Orb{Pos{3, 0}, 30}

	path := astar.AStar(start.ToString(), goal.ToString(), &astarHelper{})
	if path == nil {
		log.Fatalf("no path")
	}

	posns := []Pos{}
	fmt.Println(path)
	for i := len(path) - 1; i >= 0; i-- {
		elem := OrbFromString(path[i])
		posns = append(posns, elem.Pos)

		fmt.Println(elem)
		if i == 0 {
			continue
		}

		next := OrbFromString(path[i-1])
		//fmt.Printf("start %v, next %v\n", start, next)

		opPos := findOpPos(elem, next)
		fmt.Println(opPos)
		posns = append(posns, opPos)
	}

	for i := range posns {
		if i == len(posns)-1 {
			continue
		}

		elem := posns[i]
		next := posns[i+1]

		fmt.Print("go ")
		if elem.X < next.X {
			fmt.Println("east")
		} else if elem.X > next.X {
			fmt.Println("west")
		} else if elem.Y < next.Y {
			fmt.Println("south")
		} else {
			fmt.Println("north")
		}
	}
}
