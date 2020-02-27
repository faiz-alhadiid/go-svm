package gosvm

import (
	"gonum.org/v1/gonum/mat"
	"math"
	"math/rand"
)

// BinarySVM ...
type BinarySVM struct {
	C          float64
	Tol        float64
	MaxIter    int
	W          *mat.VecDense
	Alpha      *mat.VecDense
	B          float64
	errs       []float64
	n          int
	m          int
	data       mat.Dense
	target     []float64
	cache      *KernelCache
	kernelFunc KernelFunc
}

// NewBinarySVM ...
func NewBinarySVM(c, tol float64, maxIter int, kernelFunc KernelFunc, cache *KernelCache) *BinarySVM {
	if cache == nil {
		cache = NewKernelCache()
	}
	return &BinarySVM{
		C:          c,
		Tol:        tol,
		MaxIter:    maxIter,
		cache:      cache,
		kernelFunc: kernelFunc,
	}
}

func makeFloatArray(n int, value float64) []float64 {
	res := make([]float64, n)
	for i := 0; i < n; i++ {
		res[i] = value
	}
	return res
}

func nRange(from, to int) <-chan int {
	ch := make(chan int)
	go func() {
		for i := from; i < to; i++ {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

// SVMOut ...
func (bin *BinarySVM) SVMOut(x mat.Vector) float64 {
	return mat.Dot(bin.W, x) - bin.B
}

func (bin *BinarySVM) getKernel(i1, i2 int) float64 {
	res, err := bin.cache.Get(i1, i1)
	if err != nil {
		res = bin.kernelFunc(bin.data.RowView(i1), bin.data.RowView(i2))
		bin.cache.Add(i1, i2, res)
	}
	return res
}

func (bin *BinarySVM) takeStep(i1, i2 int) bool {
	if i1 == i2 {
		return false
	}
	a1 := bin.Alpha.AtVec(i1)
	a2 := bin.Alpha.AtVec(i2)
	y1 := bin.target[i1]
	y2 := bin.target[i2]
	E1 := bin.errs[i1]
	E2 := bin.errs[i2]
	s := y1 * y2

	var L, H float64
	if y1 != y2 {
		L = math.Max(0, a2-a1)
		H = math.Min(bin.C, bin.C+a2-a1)
	} else {
		L = math.Max(0, a1+a2-bin.C)
		H = math.Min(bin.C, a1+a2)
	}

	if L == H {
		return false
	}

	k11 := bin.getKernel(i1, i1)
	k12 := bin.getKernel(i1, i2)
	k22 := bin.getKernel(i2, i2)
	eta := k11 + k22 - 2*k12

	var a1New, a2New float64

	if eta > 0 {
		a2New = a2 + y2*(E1-E2)/eta
		if a2New <= L {
			a2New = L
		} else if a2New >= H {
			a2New = H
		}
	} else {
		f1 := y1*(E1+bin.B) - a1*k11 - s*a2*k22
		f2 := y2*(E2+bin.B) - s*a1*k11 - a2*k22
		L1 := a1 + s*(a2-L)
		H1 := a1 + s*(a2-H)
		LObj := L1*f1 + L*f2 + (L1 * L1 * k11 / 2) + (L * L * k22) + s*L*L1*k12
		HObj := H1*f1 + H*f2 + (H1 * H1 * k11 / 2) + (H * H * k22 / 2) + s*H*H1*k12
		if LObj < HObj-0.001 {
			a2New = L
		} else if LObj > HObj+0.001 {
			a2New = H
		} else {
			a2New = a2
		}
	}
	if math.Abs(a2New-a2) < 0.001*(a2New+a2+0.001) {
		return false
	}
	a1New = a1 + s*(a2-a2New)
	b1 := E1 + y1*(a1New-a1)*k11 + y2*(a2New-a2)*k12 + bin.B
	b2 := E2 + y1*(a1New-a1)*k11 + y2*(a2New-a2)*k12 + bin.B

	var bNew float64
	if 0 < a1New && a1New < bin.C {
		bNew = b1
	} else if 0 < a2New && a2New < bin.C {
		bNew = b2
	} else {
		bNew = (b1 + b2) / 2
	}

	bin.Alpha.SetVec(i1, a1New)
	bin.Alpha.SetVec(i2, a2New)
	bin.B = bNew

	var temp mat.VecDense
	add1 := y1 * (a1New - a1)
	add2 := y2 * (a2New - a2)

	var dt1, dt2 mat.VecDense
	dt1.MulElemVec(&dt1, mat.NewVecDense(bin.m, makeFloatArray(bin.m, add1)))
	dt2.MulElemVec(&dt2, mat.NewVecDense(bin.m, makeFloatArray(bin.m, add2)))
	temp.AddVec(bin.W, &dt1)
	temp.AddVec(&temp, &dt2)
	bin.W = &temp

	for i := 0; i < bin.n; i++ {
		bin.errs[i] = bin.SVMOut(bin.data.RowView(i)) - bin.target[i]
	}
	return true
}

func (bin *BinarySVM) examineExample(i2 int) int {
	y2 := bin.target[i2]
	a2 := bin.Alpha.AtVec(i2)
	err2 := bin.errs[i2]
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
			if bin.errs[i] > maxErr {
				maxErr = bin.errs[i]
				maxIndex = i
			}
			if bin.errs[i] < minErr {
				minErr = bin.errs[i]
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

	temp := make([]float64, len(target))
	for i := 0; i < bin.n; i++ {
		temp[i] = -target[i]
	}
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
