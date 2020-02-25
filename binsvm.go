package gosvm

import (
	"gonum.org/v1/gonum/mat"
	"math"
	"math/rand"
)

// BinarySVM ...
type BinarySVM struct {
	C       float64
	Tol     float64
	MaxIter int
	W       *mat.VecDense
	Alpha   *mat.VecDense
	B       float64
	errs    *mat.VecDense
	n       int
	m       int
	data    mat.Dense
	target  []float64
	cache   *KernelCache
}

// NewBinarySVM ...
func NewBinarySVM(c, tol float64, maxIter int) *BinarySVM {
	return &BinarySVM{
		C:       c,
		Tol:     tol,
		MaxIter: maxIter,
		cache:   NewKernelCache(),
	}
}

func makeFloatArray(n int, value float64) []float64 {
	res := make([]float64, n)
	for i := 0; i < n; i++ {
		res[i] = value
	}
	return res
}

func nRange(from, to int) chan int {
	ch := make(chan int)
	go func() {
		for i := from; i < to; i++ {
			ch <- i
		}
	}()
	close(ch)
	return ch
}

// SVMOut ...
func (bin *BinarySVM) SVMOut(x mat.Vector) float64 {
	return mat.Dot(bin.W, x) - bin.B
}

func (bin *BinarySVM) takeStep(i1, i2 int) bool {
	if i1 == i2 {
		return false
	}

	return false
}

func (bin *BinarySVM) examineExample(i2 int) int {
	y2 := bin.target[i2]
	a2 := bin.Alpha.AtVec(i2)
	err2 := bin.errs.AtVec(i2)
	r2 := y2 * a2

	if (r2 < bin.Tol && a2 < bin.C) || (r2 > bin.Tol && a2 > 0) {
		var i1 int
		nonZeroCCount := 0
		maxErr := math.Inf(-1)
		maxIndex := 0
		minErr := math.Inf(1)
		minIndex := 0

		nonZeroCList := []int{}
		zeroCList := []int{}

		for i := 0; i < bin.n; i++ {
			if bin.Alpha.AtVec(i) != 0 && bin.Alpha.AtVec(i) != bin.C {
				nonZeroCCount++
				nonZeroCList = append(nonZeroCList, i)
			} else {
				zeroCList = append(zeroCList, i)
			}
			if bin.errs.AtVec(i) > maxErr {
				maxErr = bin.errs.AtVec(i)
				maxIndex = i
			}
			if bin.errs.AtVec(i) < minErr {
				minErr = bin.errs.AtVec(i)
				minIndex = i
			}
		}

		if nonZeroCCount > 0 {
			if err2 > 0 {
				i1 = minIndex
			} else {
				i1 = maxIndex
			}
			step := bin.takeStep(i1, i2)
			if step {
				return 1
			}
		}

		for _, idx := range rand.Perm(len(nonZeroCList)) {
			i1 = nonZeroCList[idx]
			step := bin.takeStep(i1, i2)
			if step {
				return 1
			}
		}

		for _, idx := range rand.Perm(len(zeroCList)) {
			i1 = zeroCList[idx]
			step := bin.takeStep(i1, i2)
			if step {
				return 1
			}
		}

	}
	return 0
}

// Train ...
func (bin *BinarySVM) Train(dataTrain mat.Dense, target []float64) {
	bin.n, bin.m = dataTrain.Dims()
	bin.W = mat.NewVecDense(bin.m, nil)
	bin.Alpha = mat.NewVecDense(bin.n, nil)
	bin.target = target

	temp := mat.NewVecDense(bin.n, target)
	temp.SubVec(temp, mat.NewVecDense(bin.n, makeFloatArray(bin.n, 1)))
	bin.errs = temp

	examineAll := true
	numChanged := 0
	it := 0
	for (examineAll || numChanged < 0) && it < bin.MaxIter {
		numChanged = 0
		if examineAll {
			for i := range nRange(0, bin.n) {
				numChanged += bin.examineExample(i)
			}
		} else {
			for i := range nRange(0, bin.n) {
				alph := bin.Alpha.AtVec(i)
				if alph != 0 && alph != bin.C {
					numChanged += bin.examineExample(i)
				}
			}
		}

		if examineAll {
			examineAll = false
		} else if numChanged == 0 {
			examineAll = true
		}
		it++
	}
}
