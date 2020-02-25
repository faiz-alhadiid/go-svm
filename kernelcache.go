package gosvm

import "fmt"

// KernelCache ...
type KernelCache struct {
	diff   map[int]map[int]float64
	same   map[int]float64
	mapper func(int) int
}

// NewKernelCache ...
func NewKernelCache() *KernelCache {
	return &KernelCache{
		diff: make(map[int]map[int]float64),
		same: make(map[int]float64),
		mapper: func(a int) int {
			return a
		},
	}
}

// Add ...
func (kc *KernelCache) Add(i, j int, value float64) {
	i, j = kc.mapper(i), kc.mapper(j)
	if i == j {
		kc.same[i] = value
		return
	}
	if j < i {
		i, j = j, i
	}
	inner := kc.diff[i]
	if inner == nil {
		inner = make(map[int]float64)
	}
	inner[j] = value
	kc.diff[i] = inner

}

// Get ...
func (kc *KernelCache) Get(i, j int) (float64, error) {
	i, j = kc.mapper(i), kc.mapper(j)
	if i == j {
		v, ok := kc.same[i]
		if !ok {
			return 0.0, fmt.Errorf("Value not exist")
		}
		return v, nil
	}
	if j < i {
		i, j = j, i
	}
	inner, ok := kc.diff[i]
	if !ok {
		return 0.0, fmt.Errorf("Value not exist")
	}
	val, ok := inner[j]
	if !ok {
		return 0.0, fmt.Errorf("Value not exist")
	}
	return val, nil

}

// SliceCopy ...
func (kc *KernelCache) SliceCopy(idx []int) *KernelCache {
	mapper := func(a int) int {
		a = kc.mapper(a)
		return idx[a]
	}
	return &KernelCache{
		same:   kc.same,
		diff:   kc.diff,
		mapper: mapper,
	}
}
