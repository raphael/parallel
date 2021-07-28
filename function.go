package parallel

import "sync"

type (
	// Function provides a simple abstraction for running a job (function)
	// multiple times in parallel.
	Function struct {
		fn            WorkerFn
		n             int
		input, output chan interface{}
		errch         chan error
		wg            *sync.WaitGroup

		lock       sync.Mutex
		stopped    bool
		resHandler func(interface{})
		errHandler func(error)
	}
)

// Run creates a new function that can run fn up to n times in parallel.
func Run(fn WorkerFn, n int) *Function {
	f := &Function{
		fn:         fn,
		n:          n,
		resHandler: dummyResHandler,
		errHandler: dummyErrHandler,
	}
	f.Reset()
	return f
}

// Call schedules a single function call. Call panics if called after Wait
// unless Reset is called.
func (f *Function) Call(arg interface{}) {
	f.input <- arg
}

// Wait stops the function gracefully and returns after all pending calls have
// completed. Wait is idempotent and can be called multiple times.
func (f *Function) Wait() {
	f.lock.Lock()
	defer f.lock.Unlock()
	if f.stopped {
		return
	}
	f.stopped = true
	close(f.input)
	f.wg.Wait()
}

// Reset makes a stopped function available again for Call. The behavior of
// Reset is undefined when called on a function that has not been stopped.
func (f *Function) Reset() {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.stopped = false
	f.input = make(chan interface{})
	f.output, f.errch = Do(f.fn, f.n, f.input)
	f.wg = &sync.WaitGroup{}
	f.wg.Add(1)
	go f.processResults()
}

// OnResult provides a result handler for the function if needed. OnResult
// should be called before Call to avoid race conditions. Successive calls to
// OnResult replace the previous handler.
//
// All calls to all handlers are guaranteed to be made within a single
// goroutine. While this simplifies their implementation it also means that they
// have the potential to block the execution of the function so special care
// must be taken to make them efficient.
func (f *Function) OnResult(fn func(interface{})) { f.resHandler = fn }

// OnError provides a error handler for the function if needed. OnError should
// be called before Call to avoid race conditions. Successive calls to OnError
// replace the previous handler.
//
// All calls to all handlers are guaranteed to be made within a single
// goroutine. While this simplifies their implementation it also means that they
// have the potential to block the execution of the function so special care
// must be taken to make them efficient.
func (f *Function) OnError(fn func(error)) { f.errHandler = fn }

// processResults reads from the output and error channels and calls the
// appropriate handlers.
func (f *Function) processResults() {
	defer f.wg.Done()
	var otherClosed bool
	for {
		select {
		case res, ok := <-f.output:
			if !ok {
				if otherClosed {
					return
				}
				otherClosed = true
				continue
			}
			f.resHandler(res)

		case err, ok := <-f.errch:
			if !ok {
				if otherClosed {
					return
				}
				otherClosed = true
				continue
			}
			f.errHandler(err)
		}
	}
}

func dummyResHandler(res interface{}) {}
func dummyErrHandler(err error)       {}
