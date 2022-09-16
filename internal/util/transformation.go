package util

import (
	"sort"
	"sync"
)

type Transform[Input any, Output any] func(Input) (Output, error)

func (f Transform[Input, Output]) Do(inputs []Input) []Output {
	return do(inputs, f,
		func(i int, output Output) Output { return output },
		func(outputs []Output) []Output { return outputs })
}

type indexed[T any] struct {
	v T
	i int
}

func wrapIndexed[T any](i int, v T) indexed[T] {
	return indexed[T]{
		v: v,
		i: i,
	}
}

func unwrapIndexed[T any](outputs []indexed[T]) []T {
	sort.Slice(outputs, func(i, j int) bool { return outputs[i].i < outputs[j].i })
	sorted := make([]T, len(outputs))
	for i, o := range outputs {
		sorted[i] = o.v
	}
	return sorted
}

func (f Transform[Input, Output]) Ordered(inputs []Input) []Output {
	return do(inputs, f, wrapIndexed[Output], unwrapIndexed[Output])
}

func do[Input any, Output any, X any](inputs []Input, transform Transform[Input, Output], wrap func(int, Output) X, unwrap func([]X) []Output) []Output {
	chXs := make(chan X, len(inputs))

	wg := sync.WaitGroup{}
	wg.Add(len(inputs))
	for i, input := range inputs {
		go func(i int, input Input) {
			defer wg.Done()
			output, err := transform(input)
			if err != nil {
				return
			}
			chXs <- wrap(i, output)
		}(i, input)
	}
	wg.Wait()
	close(chXs)

	xs := make([]X, 0, len(chXs))
	for x := range chXs {
		xs = append(xs, x)
	}

	result := unwrap(xs)

	return result
}
