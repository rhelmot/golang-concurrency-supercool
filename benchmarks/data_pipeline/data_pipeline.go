//A benchmark that tests how long it takes to pass data through
//a pipeline of N channels while doing a simple computation at each step.

package main

import (
	"golang.org/x/benchmarks/driver"
	"sync"
	"flag"
)

var capacity_flag = flag.Int("channel_capacity", 1, "capacity of channel buffer (default 1)")
var capacity int
var data_size_flag = flag.Int("data_size", 1, "number of ints to pass through channels (default 2)")
var data_size int

func main() {
	flag.Parse()
	capacity = *capacity_flag
	data_size = *data_size_flag
	driver.Main("DataPipeline", benchmark)
}

func benchmark() driver.Result {
	return driver.Benchmark(benchmarkN)
}

func add1(input chan int) chan int {
	output := make(chan int, capacity)
	go func(input chan int, output chan int){
		for i := range input {
			output <- i + 1
		}
		close(output)
	}(input, output)
	return output
}

func subtract5(input chan int) chan int {
	output := make(chan int, capacity)
	go func(input chan int, output chan int){
		for i := range input {
			output <- i - 5
		}
		close(output)
	}(input, output)
	return output
}

func multiplyBy3(input chan int) chan int {
	output := make(chan int, capacity)
	go func(input chan int, output chan int){
		for i := range input {
			output <- i * 3
		}
		close(output)
	}(input, output)
	return output
}

func chainFuncNTimes(input chan int, n int, f func(input chan int) chan int) chan int {
	if (n == 1) {
		return f(input)
	} else {
		return f(chainFuncNTimes(input, n - 1, f))
	}
}

func benchmarkN(n uint64) {
	input := make(chan int, capacity)

	output := chainFuncNTimes(input, int(n), add1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func(){
		for i := range output {
			_ = i
		}
		defer wg.Done()
	}()

	for i := 0; i < data_size; i++ {
		input <- i
	}
	close(input)
	wg.Wait()
}