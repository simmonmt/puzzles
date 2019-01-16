package instruction

import (
	"fmt"
	"os"
	"strconv"
	"unicode"

	"memory"
	"reader"
	"register"
	"symtab"
)

type Context struct {
	RAM     *memory.RAM
	RegFile *register.File
	Stack   *memory.Stack
	Verbose bool
}

type CB struct {
	Hlt bool
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

func regNum(val uint16) uint {
	return uint(val - 32767)
}

func regName(num uint) string {
	return fmt.Sprintf("r%d", num)
}

func argToStr(val uint16) string {
	if isReg(val) {
		return regName(regNum(val))
	} else {
		return fmt.Sprint(val)
	}
}

func tgtToStr(val uint16, st symtab.SymTab) string {
	if isReg(val) {
		return regName(regNum(val))
	} else if ent, found := st.LookupAddr(uint(val)); found {
		return fmt.Sprintf("%d <%s>", val, ent.OffStr(uint(val)))
	}
	return strconv.Itoa(int(val))
}

func regOrVal(num uint16, regFile *register.File) uint16 {
	if isReg(num) {
		return regFile.Get(regNum(num))
	} else {
		return num
	}
}

type Inst interface {
	ToString(st symtab.SymTab) string
	Exec(ctx *Context, cb *CB)
}

type add struct {
	a, b, c uint16
}

func (i *add) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("add %v, %v, %v",
		argToStr(i.a), argToStr(i.b), argToStr(i.c))
}

func (i *add) Exec(ctx *Context, cb *CB) {
	b := regOrVal(i.b, ctx.RegFile)
	c := regOrVal(i.c, ctx.RegFile)
	a := (b + c) % 32768
	ctx.RegFile.Set(regNum(i.a), a)
}

type and struct {
	a, b, c uint16
}

func (i *and) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("and %v, %v, %v",
		argToStr(i.a), argToStr(i.b), argToStr(i.c))
}

func (i *and) Exec(ctx *Context, cb *CB) {
	ctx.RegFile.Set(regNum(i.a), regOrVal(i.b, ctx.RegFile)&regOrVal(i.c, ctx.RegFile))
}

type call struct {
	a uint16
}

func (i *call) ToString(st symtab.SymTab) string {
	return "call " + tgtToStr(i.a, st)
}

func (i *call) Exec(ctx *Context, cb *CB) {
	ctx.Stack.Push(cb.NPC)
	cb.NPC = regOrVal(i.a, ctx.RegFile)
}

type eq struct {
	a, b, c uint16
}

func (i *eq) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("eq %v, %v, %v",
		argToStr(i.a), argToStr(i.b), argToStr(i.c))
}

func (i *eq) Exec(ctx *Context, cb *CB) {
	var res uint16
	if regOrVal(i.b, ctx.RegFile) == regOrVal(i.c, ctx.RegFile) {
		res = 1
	}
	ctx.RegFile.Set(regNum(i.a), res)
}

type gt struct {
	a, b, c uint16
}

func (i *gt) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("gt %v, %v, %v",
		argToStr(i.a), argToStr(i.b), argToStr(i.c))
}

func (i *gt) Exec(ctx *Context, cb *CB) {
	var res uint16
	if regOrVal(i.b, ctx.RegFile) > regOrVal(i.c, ctx.RegFile) {
		res = 1
	}
	ctx.RegFile.Set(regNum(i.a), res)
}

type in struct {
	a uint16
}

func (i *in) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("in %v", argToStr(i.a))
}

func (i *in) Exec(ctx *Context, cb *CB) {
	b := [1]byte{}
	n, err := os.Stdin.Read(b[:])
	if n == 0 || err != nil {
		panic("bad read")
	}

	ctx.RegFile.Set(regNum(i.a), uint16(b[0]))
}

type hlt struct{}

func (i *hlt) ToString(st symtab.SymTab) string { return "hlt" }

func (i *hlt) Exec(ctx *Context, cb *CB) {
	cb.Hlt = true
}

type jmp struct {
	a uint16
}

func (i *jmp) ToString(st symtab.SymTab) string {
	return "jmp " + tgtToStr(i.a, st)
}

func (i *jmp) Exec(ctx *Context, cb *CB) {
	cb.NPC = regOrVal(i.a, ctx.RegFile)
}

type jt struct {
	cond, tgt uint16
}

func (i *jt) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("jt %s %s", argToStr(i.cond), tgtToStr(i.tgt, st))
}

func (i *jt) Exec(ctx *Context, cb *CB) {
	if regOrVal(i.cond, ctx.RegFile) == 0 {
		return
	}

	cb.NPC = regOrVal(i.tgt, ctx.RegFile)
}

type jf struct {
	cond, tgt uint16
}

func (i *jf) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("jf %s %s", argToStr(i.cond), tgtToStr(i.tgt, st))
}

func (i *jf) Exec(ctx *Context, cb *CB) {
	if regOrVal(i.cond, ctx.RegFile) != 0 {
		return
	}

	cb.NPC = regOrVal(i.tgt, ctx.RegFile)
}

type mod struct {
	a, b, c uint16
}

func (i *mod) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("mod %v, %v, %v",
		argToStr(i.a), argToStr(i.b), argToStr(i.c))
}

func (i *mod) Exec(ctx *Context, cb *CB) {
	b := regOrVal(i.b, ctx.RegFile)
	c := regOrVal(i.c, ctx.RegFile)
	a := (b % c) % 32768
	ctx.RegFile.Set(regNum(i.a), a)
}

type mult struct {
	a, b, c uint16
}

func (i *mult) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("mult %v, %v, %v",
		argToStr(i.a), argToStr(i.b), argToStr(i.c))
}

func (i *mult) Exec(ctx *Context, cb *CB) {
	b := regOrVal(i.b, ctx.RegFile)
	c := regOrVal(i.c, ctx.RegFile)
	a := (b * c) % 32768
	ctx.RegFile.Set(regNum(i.a), a)
}

type nop struct{}

func (i *nop) ToString(st symtab.SymTab) string { return "nop" }
func (i *nop) Exec(ctx *Context, cb *CB)        {}

type not struct {
	a, b uint16
}

func (i *not) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("not %v, %v", argToStr(i.a), argToStr(i.b))
}

func (i *not) Exec(ctx *Context, cb *CB) {
	ctx.RegFile.Set(regNum(i.a), (^regOrVal(i.b, ctx.RegFile))&0x7fff)
}

type or struct {
	a, b, c uint16
}

func (i *or) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("or %v, %v, %v",
		argToStr(i.a), argToStr(i.b), argToStr(i.c))
}

func (i *or) Exec(ctx *Context, cb *CB) {
	ctx.RegFile.Set(regNum(i.a), regOrVal(i.b, ctx.RegFile)|regOrVal(i.c, ctx.RegFile))
}

type ret struct{}

func (i *ret) ToString(st symtab.SymTab) string { return "ret" }

func (i *ret) Exec(ctx *Context, cb *CB) {
	dest, found := ctx.Stack.Pop()
	if !found {
		cb.Hlt = true
		return
	}

	cb.NPC = dest
}

type rmem struct {
	a, b uint16
}

func (i *rmem) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("rmem %v, %v", argToStr(i.a), tgtToStr(i.b, st))
}

func (i *rmem) Exec(ctx *Context, cb *CB) {
	val := ctx.RAM.Read(regOrVal(i.b, ctx.RegFile))
	ctx.RegFile.Set(regNum(i.a), val)
}

type out struct {
	a uint16
}

func (i *out) ToString(st symtab.SymTab) string {
	if isReg(i.a) {
		return fmt.Sprintf("out %s", regName(regNum(i.a)))
	}

	b := byte(i.a)
	if !unicode.IsPrint(rune(b)) {
		return fmt.Sprintf("out 0x%02x", b)
	}

	return fmt.Sprintf("out %s", string(b))
}

func (i *out) Exec(ctx *Context, cb *CB) {
	val := regOrVal(i.a, ctx.RegFile)
	if ctx.Verbose {
		fmt.Printf("=== out: %s (%d) ===\n", string(byte(val)), val)
	} else {
		fmt.Print(string(byte(val)))
	}
}

type pop struct {
	a uint16
}

func (i *pop) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("pop %s", argToStr(i.a))
}

func (i *pop) Exec(ctx *Context, cb *CB) {
	val, found := ctx.Stack.Pop()
	if !found {
		panic("empty stack")
	}
	ctx.RegFile.Set(regNum(i.a), val)
}

type push struct {
	a uint16
}

func (i *push) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("push %s", argToStr(i.a))
}

func (i *push) Exec(ctx *Context, cb *CB) {
	ctx.Stack.Push(regOrVal(i.a, ctx.RegFile))
}

type set struct {
	res, src uint16
}

func (i *set) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("set %s %s", argToStr(i.res), argToStr(i.src))
}

func (i *set) Exec(ctx *Context, cb *CB) {
	ctx.RegFile.Set(regNum(i.res), regOrVal(i.src, ctx.RegFile))
}

type wmem struct {
	a, b uint16
}

func (i *wmem) ToString(st symtab.SymTab) string {
	return fmt.Sprintf("wmem %v, %v", tgtToStr(i.a, st), argToStr(i.b))
}

func (i *wmem) Exec(ctx *Context, cb *CB) {
	addr := regOrVal(i.a, ctx.RegFile)
	val := regOrVal(i.b, ctx.RegFile)
	if ctx.Verbose {
		fmt.Printf("writing %v to %v\n", val, addr)
	}
	ctx.RAM.Write(addr, val)
}

func Read(sr reader.Short) (Inst, int, error) {
	op, err := sr.Read()
	if err != nil {
		return nil, 0, err
	}

	inst, argLen, err := new(op, sr)
	if err != nil {
		return nil, 1, err
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
			return nil, 0, fmt.Errorf("non-reg result")
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
			return nil, 0, fmt.Errorf("non-reg result")
		}
		return &pop{a}, 1, nil

	case 4:
		a, b, c, err := read3(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad eq read: %v", err)
		}
		if !isReg(a) {
			return nil, 0, fmt.Errorf("non-reg result")
		}
		return &eq{a, b, c}, 3, nil

	case 5:
		a, b, c, err := read3(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad gt read: %v", err)
		}
		if !isReg(a) {
			return nil, 0, fmt.Errorf("non-reg result")
		}
		return &gt{a, b, c}, 3, nil

	case 6:
		a, err := sr.Read()
		if err != nil {
			return nil, 0, fmt.Errorf("bad jmp read: %v", err)
		}
		if isReg(a) {
			panic("unexpected jmp to reg value")
		}
		return &jmp{a}, 1, nil

	case 7:
		cond, tgt, err := read2(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad jt read: %v", err)
		}
		if isReg(tgt) {
			panic("unexpected jt to tgt value")
		}
		return &jt{cond, tgt}, 2, nil

	case 8:
		cond, tgt, err := read2(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad jf read: %v", err)
		}
		if isReg(tgt) {
			panic("unexpected jf to tgt value")
		}
		return &jf{cond, tgt}, 2, nil

	case 9:
		a, b, c, err := read3(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad add read: %v", err)
		}
		if !isReg(a) {
			return nil, 0, fmt.Errorf("non-reg result")
		}
		return &add{a, b, c}, 3, nil

	case 10:
		a, b, c, err := read3(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad mult read: %v", err)
		}
		if !isReg(a) {
			return nil, 0, fmt.Errorf("non-reg result")
		}
		return &mult{a, b, c}, 3, nil

	case 11:
		a, b, c, err := read3(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad mod read: %v", err)
		}
		if !isReg(a) {
			return nil, 0, fmt.Errorf("non-reg result")
		}
		return &mod{a, b, c}, 3, nil

	case 12:
		a, b, c, err := read3(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad type or and read: %v", err)
		}
		if !isReg(a) {
			return nil, 0, fmt.Errorf("non-reg result")
		}
		return &and{a, b, c}, 3, nil

	case 13:
		a, b, c, err := read3(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad or read: %v", err)
		}
		if !isReg(a) {
			return nil, 0, fmt.Errorf("non-reg result")
		}
		return &or{a, b, c}, 3, nil

	case 14:
		a, b, err := read2(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad not read: %v", err)
		}
		if !isReg(a) {
			return nil, 0, fmt.Errorf("non-reg result")
		}
		return &not{a, b}, 2, nil

	case 15:
		a, b, err := read2(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad rmem read: %v", err)
		}
		if !isReg(a) {
			return nil, 0, fmt.Errorf("non-reg result")
		}
		return &rmem{a, b}, 2, nil

	case 16:
		a, b, err := read2(sr)
		if err != nil {
			return nil, 0, fmt.Errorf("bad wmem read: %v", err)
		}
		return &wmem{a, b}, 2, nil

	case 17:
		a, err := sr.Read()
		if err != nil {
			return nil, 0, fmt.Errorf("bad call read: %v", err)
		}
		return &call{a}, 1, nil

	case 18:
		return &ret{}, 0, nil

	case 19:
		a, err := sr.Read()
		if err != nil {
			return nil, 0, fmt.Errorf("bad out read: %v", err)
		}
		return &out{a}, 1, nil

	case 20:
		a, err := sr.Read()
		if err != nil {
			return nil, 0, fmt.Errorf("bad in read: %v", err)
		}
		if !isReg(a) {
			return nil, 0, fmt.Errorf("non-reg result")
		}
		return &in{a}, 1, nil

	case 21:
		return &nop{}, 0, nil
	default:
		return nil, 0, fmt.Errorf("unknown op %v", op)
	}
}
