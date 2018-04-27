//A sample program meant to generate a large amount of Go routines
//for experimenting with the runtime schedule tracer.

//To run:
//go build runtime_debugger_example.go
//GOMAXPROCS=2 GODEBUG=schedtrace=1000 ./runtime_debugger_example

package main

import (
    "fmt"
    "time"
)

func fibonacci(n int, result chan int) {
    var a, b int
    c := make(chan int)
    d := make(chan int)

    if (n == 1 || n == 2) {
        result <- 1
        return
    }

    go fibonacci(n - 1, c)

    go fibonacci(n - 2, d)

    for i := 0; i < 2; i++ {
        select {
        case a = <-c:
        case b = <-d:
        }
    }

    result <- (a + b)
}

func main() {
    n := 35
    result := make(chan int)
    start_time := time.Now()

    go fibonacci(n, result)

    nth_number := <-result

    fmt.Println(
        "It took",
        time.Since(start_time),
        "milliseconds to compute the",
        n,
        "Fibonacci number:",
        nth_number,
    )
}