package util

import (
	"sort"
	"sync"
)

type Transformation[Input any, Output any] struct {
	Transform func(Input) (Output, error)
	Success   *func(Input, Output)
	Failure   *func(Input, error)
}

func (t Transformation[Input, Output]) Do(inputs []Input) []Output {
	return doer(inputs, t, func(i int, output Output) Output {
		return output
	}, func(outputs []Output) []Output {
		return outputs
	})
}

type sortable[T any] struct {
	o T
	i int
}

func (t Transformation[Input, Output]) Ordered(inputs []Input) []Output {
	return doer(inputs, t, func(i int, output Output) sortable[Output] {
		return sortable[Output]{
			o: output,
			i: i,
		}
	}, func(outputs []sortable[Output]) []Output {
		sort.Slice(outputs, func(i, j int) bool { return outputs[i].i < outputs[j].i })
		sorted := make([]Output, len(outputs))
		for i, o := range outputs {
			sorted[i] = o.o
		}
		return sorted
	})
}

func doer[Input any, Output any, X any](inputs []Input, t Transformation[Input, Output], produce func(int, Output) X, process func([]X) []Output) []Output {
	chXs := make(chan X, len(inputs))

	wg := sync.WaitGroup{}
	wg.Add(len(inputs))
	for i, input := range inputs {
		go func(i int, input Input) {
			defer wg.Done()
			output, err := t.Transform(input)
			if err != nil {
				if t.Failure != nil {
					(*t.Failure)(input, err)
				}
				return
			}
			chXs <- produce(i, output)
			if t.Success != nil {
				(*t.Success)(input, output)
			}
		}(i, input)
	}
	wg.Wait()
	close(chXs)

	xs := make([]X, 0, len(chXs))
	for x := range chXs {
		xs = append(xs, x)
	}

	processed := process(xs)

	return processed
}
