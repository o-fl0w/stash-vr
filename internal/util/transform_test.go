package util

import (
	"errors"
	"reflect"
	"strconv"
	"testing"
)

func nonZeroIntToString(input int) (string, error) {
	if input == 0 {
		return "", errors.New("zero int")
	}
	return strconv.Itoa(input), nil
}

func TestTransform_Ordered(t *testing.T) {
	type args struct {
		inputs []int
	}
	tests := []struct {
		name string
		f    Transform[int, string]
		args args
		want []string
	}{
		{
			name: "verify order and nil skip",
			f:    nonZeroIntToString,
			args: args{inputs: []int{10, 0, 1, 2}},
			want: []string{"10", "1", "2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Ordered(tt.args.inputs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ordered() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkTransform_Ordered(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Transform[int, string](nonZeroIntToString).Ordered([]int{10, 0, 1, 2, 10, 0, 1, 2, 10, 0, 1, 2, 10, 0, 1, 2, 10, 0, 1, 2, 10, 0, 1, 2})
	}
}
