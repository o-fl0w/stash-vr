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

type sortable[T any] struct {
	o T
	i int
}

func (t Transformation[Input, Output]) Do(inputs []Input) []Output {
	c := make(chan Output)
	done := make(chan any)
	outputs := make([]Output, 0, len(inputs))

	go func() {
		for {
			o, ok := <-c
			if !ok {
				close(done)
				return
			}
			outputs = append(outputs, o)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(len(inputs))
	for _, input := range inputs {
		go func(input Input) {
			defer wg.Done()
			output, err := t.Transform(input)
			if err != nil {
				if t.Failure != nil {
					(*t.Failure)(input, err)
				}
				return
			}
			c <- output
			if t.Success != nil {
				(*t.Success)(input, output)
			}
		}(input)
	}
	wg.Wait()
	close(c)
	<-done
	return outputs
}

func (t Transformation[Input, Output]) Ordered(inputs []Input) []Output {
	c := make(chan sortable[Output])
	done := make(chan any)
	outputs := make([]sortable[Output], 0, len(inputs))

	go func() {
		for {
			o, ok := <-c
			if !ok {
				close(done)
				return
			}
			outputs = append(outputs, o)
		}
	}()

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
			c <- sortable[Output]{
				o: output,
				i: i,
			}
			if t.Success != nil {
				(*t.Success)(input, output)
			}
		}(i, input)
	}
	wg.Wait()
	close(c)
	<-done
	sort.Slice(outputs, func(i, j int) bool { return outputs[i].i < outputs[j].i })
	sorted := make([]Output, len(outputs))
	for i, o := range outputs {
		sorted[i] = o.o
	}
	return sorted
}
