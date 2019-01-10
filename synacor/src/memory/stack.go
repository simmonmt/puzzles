package memory

type Stack struct {
	cells []uint16
	cur   int
}

func NewStack() *Stack {
	return &Stack{
		cells: []uint16{},
		cur:   -1,
	}
}

func (s *Stack) Push(val uint16) {
	next := s.cur + 1
	if next == len(s.cells) {
		s.cells = append(s.cells, val)
	} else {
		s.cells[next] = val
	}
	s.cur = next
}

func (s *Stack) Pop() (uint16, bool) {
	if s.cur == -1 {
		return 0, false
	}

	val := s.cells[s.cur]
	s.cur--
	return val, true
}
