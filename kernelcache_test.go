package gosvm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewKernelCache(t *testing.T) {
	t.Run("Test KernelCache Created", func(t_ *testing.T) {
		cache := NewKernelCache()
		assert.NotNil(t_, cache)
	})
	t.Run("Test Each Fields", func(t_ *testing.T) {
		cache := NewKernelCache()
		assert.NotNil(t_, cache.same)
		assert.NotNil(t_, cache.diff)
		exp := 10
		assert.Equal(t_, cache.mapper(10), exp)
	})
}

func TestGet(t *testing.T) {
	t.Run("Same i,j not exist", func(t_ *testing.T) {
		cache := NewKernelCache()
		assert.Equal(t_, len(cache.same), 0)
		_, err := cache.Get(1, 1)
		assert.Error(t_, err)
	})

	t.Run("Different i,j not exist", func(t_ *testing.T) {
		cache := NewKernelCache()
		assert.Equal(t_, len(cache.diff), 0)
		_, err := cache.Get(1, 0)
		assert.Error(t_, err)
	})

	t.Run("Same i,j exist", func(t_ *testing.T) {
		cache := NewKernelCache()
		cache.same = map[int]float64{
			5: 0.25,
		}
		res, err := cache.Get(5, 5)
		assert.Nil(t_, err)
		assert.Equal(t_, res, 0.25)
	})
	t.Run("Different i, j exist", func(t_ *testing.T) {
		cache := NewKernelCache()
		cache.diff = map[int]map[int]float64{
			1: map[int]float64{
				2: 0.5,
			},
		}
		res, err := cache.Get(1, 2)
		assert.Nil(t_, err)
		assert.Equal(t_, res, 0.5)

		res, err = cache.Get(2, 1)
		assert.Nil(t_, err)
		assert.Equal(t_, res, 0.5)
	})
	t.Run("i exist but j not exist", func(t_ *testing.T) {
		cache := NewKernelCache()
		cache.diff = map[int]map[int]float64{
			1: map[int]float64{
				2: 0.5,
			},
		}

		_, err := cache.Get(1, 3)
		assert.Error(t_, err)
	})
}

func TestAdd(t *testing.T) {
	t.Run("Same i,j", func(t_ *testing.T) {
		cache := NewKernelCache()
		cache.Add(3, 3, 1.5)
		assert.Equal(t_, cache.same, map[int]float64{3: 1.5})
	})
	t.Run("Different i and j. j not exist yet ", func(t_ *testing.T) {
		cache := NewKernelCache()
		cache.Add(2, 1, 1.5)
		assert.Equal(t_, cache.diff, map[int]map[int]float64{
			1: map[int]float64{2: 1.5},
		})
	})
}

func TestSlice(t *testing.T) {
	t.Run("Test mapper function", func(t_ *testing.T) {
		cache := NewKernelCache()
		idx := []int{4, 1, 2, 5, 6, 8}
		cache2 := cache.Slice(idx)
		res := []int{}
		for i := range idx {
			res = append(res, cache2.mapper(i))
		}
		assert.Equal(t_, res, idx)
	})
	t.Run("Slice run as expected", func(t_ *testing.T) {
		original := [][]float64{
			{1, 16, 47, 72, 77},
			{16, 2, 37, 57, 65},
			{47, 37, 3, 40, 30},
			{72, 57, 40, 4, 31},
			{77, 65, 30, 31, 5},
		}
		cache := NewKernelCache()
		for i := 0; i < 5; i++ {
			for j := 0; j < 5; j++ {
				cache.Add(i, j, original[i][j])
			}
		}
		idx := []int{4, 2, 0}
		slice := cache.Slice(idx)
		expected := [][]float64{
			{5, 30, 77},
			{30, 3, 47},
			{77, 47, 1},
		}
		result := make([][]float64, 3)
		for i := range result {
			result[i] = make([]float64, 3)
		}
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				result[i][j], _ = slice.Get(i, j)
			}
		}
		assert.Equal(t_, expected, result)

	})
}

func TestAllSlicesSame(t *testing.T) {
	t.Run("All Cache slices should have same references", func(t_ *testing.T) {
		cache := NewKernelCache()
		cache.Add(1, 1, 5)
		cache.Add(2, 1, 6)
		slices := []*KernelCache{}
		for i := 0; i < 5; i++ {
			slices = append(slices, cache.Slice([]int{}))
		}
		for i := 0; i < 5; i++ {
			assert.Equal(t_, slices[i].diff, cache.diff)
			assert.Equal(t_, slices[i].same, cache.same)
			assert.Equal(t_, slices[i].mut, cache.mut)
		}
	})

	t.Run("Any mutation in each slices should modify the original and other slices", func(t_ *testing.T) {
		original := [][]float64{
			{1, 16, 47, 72, 77},
			{16, 2, 37, 57, 65},
			{47, 37, 3, 40, 30},
			{72, 57, 40, 4, 31},
			{77, 65, 30, 31, 5},
		}
		cache := NewKernelCache()
		for i := 0; i < 5; i++ {
			for j := 0; j < 5; j++ {
				cache.Add(i, j, original[i][j])
			}
		}
		slices := make([]*KernelCache, 3)
		cop := [][]int{
			{3, 2, 4},
			{1, 3, 0},
			{2, 1, 3},
		}
		for i := 0; i < 3; i++ {
			slices[i] = cache.Slice(cop[i])
		}
		slices[0].Add(1, 1, 7)
		slices[1].Add(0, 0, 5)
		for i := 0; i < 3; i++ {
			assert.Equal(t_, slices[i].diff, cache.diff)
			assert.Equal(t_, slices[i].same, cache.same)
		}
	})
}
