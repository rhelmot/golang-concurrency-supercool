//An example benchmark that tests how long it takes to pass n
//integers through a channel.

package main

import (
	"golang.org/x/benchmarks/driver"
)

func main() {
	driver.Main("EXAMPLE", benchmark)
}

func benchmark() driver.Result {
	return driver.Benchmark(benchmarkN)
}

func benchmarkN(n uint64) {
	c := make(chan int)

	go func(){
		for i := 0; i < int(n); i++ {
			c <- i
		}
		close(c)
	}()

	for i := range c {
		_ = i
	}
}