package util

import (
	"reflect"
	"strconv"
	"testing"
)

func TestTransform_Ordered(t *testing.T) {
	var nonZeroIntToString = func(input int) *string {
		if input == 0 {
			return nil
		}
		s := strconv.Itoa(input)
		return &s
	}
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
