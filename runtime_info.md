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
| G    | Goroutine    |  Dynamically growing/shrinking user stack (used for the execution of the Go code itself)     |
| M    | OS thread   |  Fixed-size system stack (and signal stack on Unix) |
| P | CPU resource | N/A |

(Source: https://github.com/golang/go/blob/master/src/runtime/HACKING.md)

G source code: https://github.com/golang/go/blob/5dd978a283ca445f8b5f255773b3904497365b61/src/runtime/runtime2.go#L332

M source code: https://github.com/golang/go/blob/5dd978a283ca445f8b5f255773b3904497365b61/src/runtime/runtime2.go#L403

P source code: https://github.com/golang/go/blob/5dd978a283ca445f8b5f255773b3904497365b61/src/runtime/runtime2.go#L470

### Behavior

The Go scheduler is [partially pre-emptive](https://github.com/golang/go/issues/11462) and work-stealing.

Every P has a queue of goroutines that are ready to be run.
There is also a global queue of goroutines.
If a P runs out of work, it will first look in the global queue.
If the global queue is empty, it will then try to steal work from another P.

An M must own a P in order to execute user Go code.
When M is not executing user code, e.g. when it is performing a syscall, it does not need a P and therefore will release it.
When the M wants to resume executing user code, it will need to reaquire a P.

A goroutine can yeild to another goroutine by calling `runtime.Gosched()`.

### Settings

The number of P can be set at execution time with the GOMAXPROCS variable. e.g.:

>GOMAXPROCS=4 ./test

Or, inside the code itself:

> runtime.GOMAXPROCS(4)

## Memory Management

### Allocation

Allocator source: https://github.com/golang/go/blob/5dd978a283ca445f8b5f255773b3904497365b61/src/runtime/malloc.go#L550

Memory allocation works differently for small and large objects.

#### Small objects

Go has a set of size classes, defined [here](https://github.com/golang/go/blob/7ee43faf78f3b0c97c315c28f13dd802047af0c8/src/runtime/mksizeclasses.go).
Every P has a cache of mspans: spans of pages which are themselves split into segments of a given size class.
Within each mspan, a bitmap is used to mark which segments are free and which have been allocated.
When a small object is allocated, Go takes the smallest size class that can fit the object and looks at the bitmap for its mspan to see if there is free memory of that size.
If so, that memory is allocated.
If not, the allocator gives the P a new mspan for that size, then allocates from there.

This method of allocation reduces fragmentation (which is important, because the garbage collector is non-compacting).

#### Large Objects

Source code for allocation of large objects: https://github.com/golang/go/blob/5dd978a283ca445f8b5f255773b3904497365b61/src/runtime/malloc.go#L779

Large objects are allocated directly from the heap.

### Garbage Collection

Go's garbage collector is concurrent mark-sweep with a write barrier. 

Garbage collector source: https://github.com/golang/go/blob/5dd978a283ca445f8b5f255773b3904497365b61/src/runtime/mgc.go