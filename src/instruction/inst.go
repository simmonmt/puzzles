package instruction

import (
	"fmt"
	"unicode"

	"memory"
	"reader"
	"register"
)

type CB struct {
	Hlt bool
	Jmp bool
	NPC uint16
}

func read2(sr reader.Short) (a, b uint16, err error) {
	a, err = sr.Read()
	if err != nil {
		return 0, 0, err
	}
	b, err = sr.Read()
	if err != nil {
		return 0, 0, err
	}
	return a, b, nil
}

func read3(sr reader.Short) (a, b, c uint16, err error) {
	a, err = sr.Read()
	if err != nil {
		return 0, 0, 0, err
	}
	b, err = sr.Read()
	if err != nil {
		return 0, 0, 0, err
	}
	c, err = sr.Read()
	if err != nil {
		return 0, 0, 0, err
	}
	return a, b, c, nil
}

func isReg(val uint16) bool {
	return val > 32767
}

func regNum(val uint16) uint8 {
	return uint8(val - 32767)
}

func argToStr(val uint16) string {
	if isReg(val) {
		return fmt.Sprintf("r%d", regNum(val))
	} else {
		return fmt.Sprint(val)
	}
}

func regOrVal(num uint16, regFile *register.File) uint16 {
	if isReg(num) {
		return regFile[regNum(num)]
	} else {
		return num
	}
}

type Inst interface {
	String() string
	Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB
}

type add struct {
	a, b, c uint16
}

func (i *add) String() string {
	return fmt.Sprintf("add %v, %v, %v",
		argToStr(i.a), argToStr(i.b), argToStr(i.c))
}

func (i *add) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	regFile[regNum(i.a)] = uint16(uint32(regOrVal(i.b, regFile)) +
		uint32(regOrVal(i.c, regFile)))
	return &CB{}
}

type and struct {
	a, b, c uint16
}

func (i *and) String() string {
	return fmt.Sprintf("and %v, %v, %v",
		argToStr(i.a), argToStr(i.b), argToStr(i.c))
}

func (i *and) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	regFile[regNum(i.a)] = regOrVal(i.b, regFile) & regOrVal(i.c, regFile)
	return &CB{}
}

type eq struct {
	a, b, c uint16
}

func (i *eq) String() string {
	return fmt.Sprintf("eq %v, %v, %v",
		argToStr(i.a), argToStr(i.b), argToStr(i.c))
}

func (i *eq) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	var res uint16
	if regOrVal(i.b, regFile) == regOrVal(i.c, regFile) {
		res = 1
	}
	regFile[regNum(i.a)] = res

	return &CB{}
}

type gt struct {
	a, b, c uint16
}

func (i *gt) String() string {
	return fmt.Sprintf("gt %v, %v, %v",
		argToStr(i.a), argToStr(i.b), argToStr(i.c))
}

func (i *gt) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	var res uint16
	if regOrVal(i.b, regFile) > regOrVal(i.c, regFile) {
		res = 1
	}
	regFile[regNum(i.a)] = res

	return &CB{}
}

type hlt struct{}

func (i *hlt) String() string { return "hlt" }

func (i *hlt) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	return &CB{Hlt: true}
}

type jmp struct {
	val uint16
}

func (i *jmp) String() string {
	return fmt.Sprintf("jmp %s+1", argToStr(i.val))
}

func (i *jmp) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	if isReg(i.val) {
		panic("reg jmp unimplemented")
	}

	return &CB{
		Jmp: true,
		NPC: i.val,
	}
}

type jt struct {
	cond, tgt uint16
}

func (i *jt) String() string {
	return fmt.Sprintf("jt %s %s+1", argToStr(i.cond), argToStr(i.tgt))
}

func (i *jt) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	if regOrVal(i.cond, regFile) == 0 {
		return &CB{}
	}

	return &CB{
		Jmp: true,
		NPC: regOrVal(i.tgt, regFile),
	}
}

type jf struct {
	cond, tgt uint16
}

func (i *jf) String() string {
	return fmt.Sprintf("jf %s %s+1", argToStr(i.cond), argToStr(i.tgt))
}

func (i *jf) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	if regOrVal(i.cond, regFile) != 0 {
		return &CB{}
	}

	return &CB{
		Jmp: true,
		NPC: regOrVal(i.tgt, regFile),
	}
}

type nop struct{}

func (i *nop) String() string { return "nop" }

func (i *nop) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	return &CB{}
}

type not struct {
	a, b uint16
}

func (i *not) String() string {
	return fmt.Sprintf("not %v, %v", argToStr(i.a), argToStr(i.b))
}

func (i *not) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	regFile[regNum(i.a)] = (^regOrVal(i.b, regFile)) & 0x7fff
	return &CB{}
}

type or struct {
	a, b, c uint16
}

func (i *or) String() string {
	return fmt.Sprintf("or %v, %v, %v",
		argToStr(i.a), argToStr(i.b), argToStr(i.c))
}

func (i *or) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	regFile[regNum(i.a)] = regOrVal(i.b, regFile) | regOrVal(i.c, regFile)
	return &CB{}
}

type set struct {
	res, src uint16
}

func (i *set) String() string {
	return fmt.Sprintf("set %s %s", argToStr(i.res), argToStr(i.src))
}

func (i *set) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	regFile[regNum(i.res)] = regOrVal(i.src, regFile)
	return &CB{}
}

type out struct {
	ch rune
}

func (i *out) String() string {
	v := string(i.ch)
	if !unicode.IsPrint(i.ch) {
		v = fmt.Sprintf("0x%02x", byte(i.ch))
	}

	return fmt.Sprintf("out %s", v)
}

func (i *out) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	fmt.Print(string(i.ch))
	return &CB{}
}

type pop struct {
	a uint16
}

func (i *pop) String() string {
	return fmt.Sprintf("pop %s", argToStr(i.a))
}

func (i *pop) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	val, err := stack.Pop()
	if err != nil {
		panic(err.Error())
	}
	regFile[regNum(i.a)] = val
	return &CB{}
}

type push struct {
	a uint16
}

func (i *push) String() string {
	return fmt.Sprintf("push %s", argToStr(i.a))
}

func (i *push) Exec(ram *memory.RAM, regFile *register.File, stack *memory.Stack) *CB {
	stack.Push(regOrVal(i.a, regFile))
	return &CB{}
}

func Read(sr reader.Short) (Inst, int, error) {
	op, err := sr.Read()
	if err != nil {
		return nil, 0, err
	}

	inst, argLen, err := new(op, sr)
	if err != nil {
		return nil, 0, err
	}

	return inst, argLen + 1, nil
}

func new(op uint16, sr reader.Short) (Inst, int, error) {
	switch op {
	case 0:
		return &hlt{}, 0, nil

	case 1:
		res, src, err := read2(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad set read: %v", err)
		}
		if !isReg(res) {
			panic("set with non-reg res")
		}
		return &set{res: res, src: src}, 2, nil

	case 2:
		a, err := sr.Read()
		if err != nil {
			return nil, 0, fmt.Errorf("bad push read: %v", err)
		}
		return &push{a}, 1, nil

	case 3:
		a, err := sr.Read()
		if err != nil {
			return nil, 0, fmt.Errorf("bad pop read: %v", err)
		}
		if !isReg(a) {
			panic("pop with non-reg res")
		}
		return &pop{a}, 1, nil

	case 4:
		a, b, c, err := read3(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad eq read: %v", err)
		}
		if !isReg(a) {
			panic("eq with non-reg res")
		}
		return &eq{a, b, c}, 3, nil

	case 5:
		a, b, c, err := read3(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad gt read: %v", err)
		}
		if !isReg(a) {
			panic("gt with non-reg res")
		}
		return &gt{a, b, c}, 3, nil

	case 6:
		val, err := sr.Read()
		if err != nil {
			return nil, 0, fmt.Errorf("bad jmp read: %v", err)
		}
		return &jmp{val}, 1, nil

	case 7:
		cond, tgt, err := read2(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad jt read: %v", err)
		}
		return &jt{cond, tgt}, 2, nil

	case 8:
		cond, tgt, err := read2(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad jf read: %v", err)
		}
		return &jf{cond, tgt}, 2, nil

	case 9:
		a, b, c, err := read3(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad add read: %v", err)
		}
		if !isReg(a) {
			panic("add with non-reg res")
		}
		return &add{a, b, c}, 3, nil

	case 12:
		a, b, c, err := read3(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad type or and read: %v", err)
		}
		if !isReg(a) {
			panic("and with non-reg res")
		}
		return &and{a, b, c}, 3, nil

	case 13:
		a, b, c, err := read3(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad or read: %v", err)
		}
		if !isReg(a) {
			panic("or with non-reg res")
		}
		return &or{a, b, c}, 3, nil

	case 14:
		a, b, err := read2(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad not read: %v", err)
		}
		if !isReg(a) {
			panic("not with non-reg res")
		}
		return &not{a, b}, 2, nil

	case 19:
		val, err := sr.Read()
		if err != nil {
			return nil, 0, fmt.Errorf("bad out read: %v", err)
		}
		return &out{rune(val & 0xff)}, 1, nil

	case 21:
		return &nop{}, 0, nil
	default:
		return nil, 0, fmt.Errorf("unknown op %v", op)
	}
}
