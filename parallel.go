/*
Package parallel provides a simple abstraction for running a function many times
in parallel potentially providing different values for its arguments in each
invocation. Parallel returns results as they become available making it possible
to start processing them while calls are still in progress or even pending.
*/

package parallel

import "sync"

type (
	// WorkerFn is a function that is executed in parallel.
	WorkerFn func(payload interface{}) (result interface{}, err error)
)

// Do runs n goroutines that call fn in parallel. It reads the function's
// argument from the input channel and writes the results and any error to the
// returned channels. The returned channels are closed once all inputs have been
// processed (that is the input channel is empty and closed).
func Do(fn WorkerFn, n int, input chan interface{}) (chan interface{}, chan error) {
	resch := make(chan interface{})
	errch := make(chan error)
	var wg sync.WaitGroup

	// Spawn n goroutines that consume from the input channel, call fn and
	// write the results and errors to the returned channels.
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(input, resch chan interface{}, errch chan error, wg *sync.WaitGroup) {
			defer wg.Done()
			for in := range input {
				res, err := fn(in)
				if err != nil {
					errch <- err
				} else {
					resch <- res
				}
			}
		}(input, resch, errch, &wg)
	}

	// Close result and error channels once all goroutines have finished.
	go func() {
		wg.Wait()
		close(resch)
		close(errch)
	}()

	return resch, errch
}
