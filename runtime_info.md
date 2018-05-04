# Write-up for the Go runtime

## Basics

Unlike languages like Java, Go's runtime does *not* involve a virtual machine.
In this case, 'runtime' just refers to a library that implements features of the language like garbage collection and concurrency.
(This information is from https://golang.org/doc/faq#runtime. )
The source for the runtime can be found at https://github.com/golang/go/tree/master/src/runtime.
A Go package for interacting with the runtime can be found at https://golang.org/pkg/runtime/.


## Getting Runtime Information From Running Programs

In order to get debugging information from the runtime, we set the GODEBUG variable when launching the program.
For example, `GODEBUG=schedtrace=1000 ./test` will print information once every 1000 milliseconds about how the scheduler is performing.
(How many processors are idle, how many goroutines are queued for each processor, etc.)


Setting `scheddetail` to a number of milliseconds will print more detailed information.
For it to work, `schedtrace` must also be set.
(Example: `GOMAXPROCS=2 GODEBUG=schedtrace=1000,scheddetail=1000 ./test`)

More information about the GODEBUG variable can be found at https://golang.org/pkg/runtime/.

## Scheduler

Source code: https://github.com/golang/go/blob/4d7cf3fedbc382215df5ff6167ee9782a9cc9375/src/runtime/proc.go

### Structures

All structures are heap-allocated.
These data structures are *never* freed.
Instead, they are placed into a free pool for that specific type.

| Name | Description  | Stack |
|------|--------------|-------|
| G    | Goroutine    |  Dynamically growing/shrinking user stack     |
| M    | OS thread   |  Fixed-size system stack (and signal stack on Unix) |
| P | CPU resource | N/A |

(Source: https://github.com/golang/go/blob/master/src/runtime/HACKING.md)

G source code: https://github.com/golang/go/blob/5dd978a283ca445f8b5f255773b3904497365b61/src/runtime/runtime2.go#L332

M source code: https://github.com/golang/go/blob/5dd978a283ca445f8b5f255773b3904497365b61/src/runtime/runtime2.go#L403

P source code: https://github.com/golang/go/blob/5dd978a283ca445f8b5f255773b3904497365b61/src/runtime/runtime2.go#L470


