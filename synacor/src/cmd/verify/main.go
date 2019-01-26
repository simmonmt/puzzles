package main

import (
	"flag"
	"fmt"
)

var (
	r8 = flag.Uint("r8", 32768, "r8")

	level = 0
)

func verify(a, b uint16) (uint16, uint16) {
	//fmt.Printf("%*scalled %v,%v\n", level*2, "", a, b)
	level++

	if a == 0 {
		a = (b + 1) % 32768
		goto ret
	}
	if b == 0 {
		a, b = callVerify((a+32767)%32768, uint16(*r8))
		goto ret
	}

	b, _ = callVerify(a, (b+32767)%32768)

	a, b = callVerify((a+32767)%32768, b)

ret:
	level--
	//fmt.Printf("%*sreturning %v,%v\n", level*2, "", a, b)
	return a, b
}

type Pair struct {
	a, b uint16
}

var (
	seen = map[Pair]Pair{}
)

func callVerify(a, b uint16) (uint16, uint16) {
	in := Pair{a, b}
	if ret, found := seen[in]; found {
		return ret.a, ret.b
	}

	var res Pair
	res.a, res.b = verify(a, b)
	seen[in] = res
	return res.a, res.b
}

func main() {
	flag.Parse()

	if *r8 != 32768 {
		a, b := verify(4, 1)
		fmt.Printf("r8=%d a=%d, b=%d\n", *r8, a, b)
		return
	}

	for v := 1; v < 32768; v++ {
		*r8 = uint(v)
		seen = map[Pair]Pair{}
		a, _ := verify(4, 1)

		fmt.Printf("r8=%d a=%d\n", *r8, a)
		if a == 6 {
			break
		}
	}
}
