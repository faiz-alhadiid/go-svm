package main

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
)

func main() {
	vec := mat.NewVecDense(4, nil)
	fmt.Println(vec)

}
