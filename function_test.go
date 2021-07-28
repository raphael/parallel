package parallel

import (
	"sort"
	"testing"
)

func TestFunction(t *testing.T) {
	f := Run(idemFunc, 10)
	var res []int
	f.OnResult(func(result interface{}) {
		res = append(res, result.(int))
	})
	for i := 0; i < 100; i++ {
		f.Call(i)
	}
	f.Wait()
	if len(res) != 100 {
		t.Fatalf("Got %d results, expected 100", len(res))
	}
	sort.Ints(res)
	for i := 0; i < 100; i++ {
		if res[i] != i {
			t.Errorf("Got %v, expected %v", res[i], i)
		}
	}
}

func TestFunctionError(t *testing.T) {
	f := Run(errFunc, 10)
	var errs []error
	f.OnError(func(err error) {
		errs = append(errs, err)
	})
	for i := 0; i < 100; i++ {
		f.Call(i)
	}
	f.Wait()
	if len(errs) != 100 {
		t.Fatalf("Got %d errors, expected 100", len(errs))
	}
	for i := 0; i < 100; i++ {
		if errs[i] != errTest {
			t.Errorf("Got %v, expected %v", errs[i], errTest)
		}
	}
}

func TestFunctionStopIdempotent(t *testing.T) {
	f := Run(idemFunc, 10)
	for i := 0; i < 100; i++ {
		f.Call(i)
	}
	f.Wait()
	f.Wait()
}
