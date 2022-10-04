package util

import (
	"sort"
	"sync"
)

type Transform[Input any, Output any] func(Input) *Output

func (f Transform[Input, Output]) Ordered(inputs []Input) []Output {
	chXs := make(chan indexed[Output], len(inputs))

	wg := sync.WaitGroup{}
	wg.Add(len(inputs))
	for i, input := range inputs {
		go func(i int, input Input) {
			defer wg.Done()
			output := f(input)
			if output == nil {
				return
			}
			chXs <- wrap(i, *output)
		}(i, input)
	}
	wg.Wait()
	close(chXs)

	xs := make([]indexed[Output], 0, len(chXs))
	for x := range chXs {
		xs = append(xs, x)
	}

	result := unwrap(xs)

	return result
}

type indexed[T any] struct {
	i int
	v T
}

func wrap[T any](i int, v T) indexed[T] {
	return indexed[T]{
		i: i,
		v: v,
	}
}

func unwrap[T any](outputs []indexed[T]) []T {
	sort.Slice(outputs, func(i, j int) bool { return outputs[i].i < outputs[j].i })
	sorted := make([]T, len(outputs))
	for i, o := range outputs {
		sorted[i] = o.v
	}
	return sorted
}