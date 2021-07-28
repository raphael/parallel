# Parallel
 
[![Go Reference](https://pkg.go.dev/badge/github.com/raphael/parallel.svg)](https://pkg.go.dev/github.com/raphael/parallel)
 [![MIT License](https://img.shields.io/badge/License-MIT-brightgreen.svg?style=flat-square)](https://opensource.org/licenses/MIT)

Parallel is a simple Go package for queueing parallel executions of a given
function potentially providing different values for its arguments in each
invocation.  Parallel returns results as they become available making it
possible to start processing them while calls are still in progress or even
pending.

A prototypical use case is the implementation of an HTTP API poller that needs
to make many (10,000+) API requests each with a different payload. The package
is use case agnostic though and can be used any time a function needs to be
called many times in parallel.

The pattern implemented by the package is not novel however parallel provides a
simple and clean API that is easy to use and reason about.

## Features

  - Simple API
  - Ability to provide different values for each invocation
  - Processes results as soon as they become available
  - No dependencies on 3rd party packages
  - No need to worry about concurrency
  - As efficient as it gets

## Installation

        go get github.com/raphael/parallel

## Usage

The package can be used in two ways:
  - Using the high level `Function` type
  - Using the low level `Do` function

The `Function` type methods make use of the `Do` function internally and take
care of doing the necessary channel management to expose a simple API:

```go
// Run the function fn up to n times in parallel (note: the function won't
// actually execute until Call is invoked below).
f := parallel.Run(fn, n)
// Call the function (Call can be invoked any number of times and is
// goroutine-safe).
f.Call(args)
// Wait for completion of all calls.
f.Wait()
```

Results returned by the function can be processed using the `OnResult` method:

```go
f.OnResult(func(result interface{}) {
        // Do something with the result.
})
```

Similarly errors can be handled using the `OnError` method:

```go
f.OnError(func(err error) {
        // Do something with the error.
})
```

The `Function` type is the preferred way to use the package as it provides a
cleaner API and is easier to use. However the `Do` function might be more
convenient in some cases, for example if the code already makes use of channels
to handle concurrency.

```go
// Create the input channel used to provide arguments to the function.
input := make(chan interface{})

// Run the function fn up to n times in parallel.
resch, errch := parallel.Do(fn, n, input)

// Write to input channel (typically in a separate goroutine).
go func(input chan interface{}) {
    input <- someValue // as many times as needed
    // Close input channel once we have written all values.
    close(input)
}(input)

// Read from result and error channels (typically in the main goroutine).
loop:
    for {
        select {
            case res, ok := <-resch:
                if !ok {
                    break loop // resch is closed, we are done
                }
                // do something with res
            case err, ok := <-errch:
                if ok {
                    // do something with err
                }
        }
    }
```