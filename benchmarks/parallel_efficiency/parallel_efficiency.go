//A simple benchmark for analyzing parallel efficiency

package main

import (
	"flag"
	"runtime"
	"sync"
	"time"
	"fmt"
)

var max_procs_flag = flag.Int("max_procs", runtime.NumCPU(), "maximum number of parallel logical cpus (default varies by machine)")
var max_procs int
var iterations_flag = flag.Int("iterations", 100000, "loop iterations per go routine (default 1000)")
var iterations int
var time_on_max_procs_flag = flag.Int("time_on_max_procs", 4000, "The minimum time (in ms) that the benchmark should take to run on the maximum number of processors (default 4s)")
var time_on_max_procs int

func main() {
	flag.Parse()
	max_procs = *max_procs_flag
	iterations = *iterations_flag
	time_on_max_procs = *time_on_max_procs_flag
	times := make(map[int]int64)

	nRoutines := 100

	//Find a number of goroutines that take 4 seconds on max number
	//of processors. This will make it easier to see changes in
	//parallel efficiency.
	for timeNGoRoutines(nRoutines, max_procs) / (1000 * 1000) < int64(time_on_max_procs)  {
		nRoutines *= 10
	}

	fmt.Println("About to launch", nRoutines, "goroutines")

	for nprocs := 1; nprocs <= max_procs; nprocs++ {
		times[nprocs] = timeNGoRoutines(nRoutines, nprocs)
		fmt.Println(float64(times[nprocs]) / (1000.0 * 1000 * 1000))
	}

	fmt.Println("Parallel efficiency:")
	for nprocs := 1; nprocs <= max_procs; nprocs++ {
		fmt.Println(nprocs," cores:", float64(times[1])/(float64(times[nprocs]) * float64(nprocs)))
	}

	//fmt.Println(float64(times) / (1000.0 * 1000 * 1000))
}

func timeNGoRoutines(n int, nprocs int) int64 {
	runtime.GOMAXPROCS(nprocs)
	var wg sync.WaitGroup

	start := time.Now()
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			var j int
			for k := 0; k < iterations; k++ {
				j++
			}
		}(&wg)
	}
	wg.Wait()

	return time.Since(start).Nanoseconds()

}
