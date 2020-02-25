package main

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
)

type test struct {
	x map[int]map[int]int
}

func (t *test) cop() *test {
	return &test{
		x: t.x,
	}
}

func main() {
	vec := mat.NewVecDense(4, nil)
	fmt.Println(vec)

	t := &test{}
	fmt.Println(t.x[0] == nil)
	t.x[1] = map[int]int{2: 4}
	cop := t.cop()
	fmt.Println(cop.x)
}
