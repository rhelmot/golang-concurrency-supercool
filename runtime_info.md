# Write-up for the Go runtime

## Basics

Unlike languages like Java, Go's runtime does *not* involve a virtual machine.
In this case, 'runtime' just refers to a library that implements features of the language like garbage collection and concurrency.
(This information is from https://golang.org/doc/faq#runtime. )
The source for the runtime can be found at https://github.com/golang/go/tree/master/src/runtime.
A Go package for interacting with the runtime can be found at https://golang.org/pkg/runtime/.



