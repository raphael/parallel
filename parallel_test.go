package parallel

import (
	"fmt"
	"sort"
	"sync"
	"testing"
)

func TestDo(t *testing.T) {
	ln := 100
	input := make(chan interface{})
	resch, errsch := Do(idemFunc, 10, input)
	go func(input chan interface{}) {
		for i := 0; i < ln; i++ {
			input <- i
		}
		close(input)
	}(input)
	var results []interface{}
loop:
	for {
		select {
		case res, ok := <-resch:
			if !ok {
				break loop
			}
			results = append(results, res)
		case err, ok := <-errsch:
			if ok {
				t.Error(err)
				return
			}
		}
	}
	if len(results) != ln {
		t.Errorf("Got %d results, expected %d", len(results), ln)
	}
	resints := make([]int, len(results))
	for i, res := range results {
		resints[i] = res.(int)
	}
	sort.Ints(resints)
	for i := 0; i < len(resints); i++ {
		if resints[i] != i {
			t.Errorf("Got %d, expected %d", resints[i], i)
		}
	}
}

func TestDoErr(t *testing.T) {
	input := make(chan interface{})
	resch, errsch := Do(errFunc, 10, input)
	go func(input chan interface{}) {
		input <- 0
		close(input)
	}(input)
	var errDos []error
loop:
	for {
		select {
		case res, ok := <-resch:
			if !ok {
				break loop
			}
			t.Errorf("Got %v, expected error", res)
		case err, ok := <-errsch:
			if ok {
				errDos = append(errDos, err)
			}
		}
	}
	if len(errDos) != 1 {
		t.Fatalf("Got %d errors, expected 1", len(errDos))
	}
	if errDos[0] != errTest {
		t.Errorf("Got error %v, expected %v", errDos[0], errTest)
	}
}

func TestDoParallelism(t *testing.T) {
	n := 10
	input := make(chan interface{}, n)
	var wg sync.WaitGroup
	wg.Add(n)
	resch, errsch := Do(requireParallelFunc(&wg), n, input)
	go func(input chan interface{}) {
		for i := 0; i < n; i++ {
			input <- i
		}
		close(input)
	}(input)
	var results []interface{}
loop:
	for {
		select {
		case res, ok := <-resch:
			if !ok {
				break loop
			}
			results = append(results, res)
		case err, ok := <-errsch:
			if ok {
				t.Error(err)
				return
			}
		}
	}
	if len(results) != n {
		t.Errorf("Got %d results, expected %d", len(results), n)
	}
}

func idemFunc(p interface{}) (interface{}, error) {
	return p, nil
}

var errTest = fmt.Errorf("test error")

func errFunc(p interface{}) (interface{}, error) {
	return nil, errTest
}

func requireParallelFunc(wg *sync.WaitGroup) func(interface{}) (interface{}, error) {
	return func(p interface{}) (interface{}, error) {
		wg.Done()
		wg.Wait() // Blocks until all the goroutines call Done.
		return p, nil
	}
}
