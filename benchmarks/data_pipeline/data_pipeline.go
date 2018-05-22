//A benchmark that tests how long it takes to pass data through
//a pipeline of N channels while doing a simple computation at each step.

package main

import (
	"sync"
	"flag"
	"fmt"
	"time"
)

var capacity_flag = flag.Int("channel_capacity", 1, "capacity of channel buffer (default 1)")
var capacity int
var data_size_flag = flag.Int("data_size", 1, "number of ints to pass through channels (default 2)")
var data_size int
var max_steps_flag = flag.Int("max_steps", 100000, "maximum number of steps in the pipeline (default 100000)")
var max_steps int

func main() {
	flag.Parse()
	capacity = *capacity_flag
	data_size = *data_size_flag
	max_steps = *max_steps_flag
	times := make(map[int]int64)
	fmt.Println("ms per pipeline step:")
	for pipelineLength := 100; pipelineLength <= max_steps; pipelineLength *= 2 {
		times[pipelineLength] = timePipeline(pipelineLength)
		timePerStep := float64(times[pipelineLength]) / (1000 * 1000 * float64(pipelineLength))
		fmt.Printf("%10d steps: %f\n", pipelineLength, timePerStep)
	}
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

func timePipeline(n int) int64 {
	input := make(chan int, capacity)

	output := chainFuncNTimes(input, int(n), add1)

	var wg sync.WaitGroup
	wg.Add(1)
	start := time.Now()
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

	return time.Since(start).Nanoseconds()
}