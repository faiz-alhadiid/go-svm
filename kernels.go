package gosvm

import (
	"gonum.org/v1/gonum/mat"
)

// KernelFunc ...
type KernelFunc func(mat.Vector, mat.Vector) float64

// LinearKernel ...
func LinearKernel(x1 mat.Vector, y1 mat.Vector) float64 {
	return mat.Dot(x1, y1)
}
