package gosvm

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_makeFloatArray(t *testing.T) {
	type args struct {
		n     int
		value float64
	}
	tests := []struct {
		name string
		args args
		want []float64
	}{
		{
			name: "Create Float array of same element",
			args: args{5, 0.5},
			want: []float64{0.5, 0.5, 0.5, 0.5, 0.5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeFloatArray(tt.args.n, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeFloatArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nRange(t *testing.T) {
	t.Run("Create range iterator", func(t_ *testing.T) {
		it := nRange(0, 10)
		res := []int{}
		for i := range it {
			res = append(res, i)
		}
		assert.Equal(t_, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, res)
	})
}
