//A benchmark for assessing the accuracy of timeouts when large
//numbers of goroutines are running simultaneously

package main

import (
	"flag"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
	"fmt"
	"log"
	"os"
	"strconv"
)

var max_procs_flag = flag.Int("max_procs", runtime.NumCPU(), "maximum number of parallel logical cpus (default varies by machine)")
var max_procs int
var iterations_flag = flag.Int("iterations", 1000000, "loop iterations per go routine (default 1000000)")
var iterations int
var start_goroutines_flag = flag.Int("start_goroutines", 1000, "The initial number of goroutines to launch (default 1000)")
var start_goroutines int
var step_size_flag = flag.Int("step_size", 2, "Multiplicative increase in goroutines at every step (default 10)")
var step_size int
var n_steps_flag = flag.Int("n_steps", 5, "Number of steps (default 2)")
var n_steps int
var timeout_flag = flag.Int("timeout", 500, "Timeout length in ms (default 500)")
var timeout int
var manually_yeild_flag = flag.Bool("manually_yeild", false, "Instruct the extra goroutines to yeild the scheduler every 50 loop iterations")
var manually_yeild bool

func main() {
	flag.Parse()
	max_procs = *max_procs_flag
	iterations = *iterations_flag
	step_size = *step_size_flag
	timeout = *timeout_flag
	manually_yeild = *manually_yeild_flag
	times := make(map[int]int64)

	nRoutines := *start_goroutines_flag
	for step := 0; step < *n_steps_flag; step++ {
		runtime.GC()
		times[step] = benchmarkTimeout(nRoutines, max_procs)
		fmt.Println(nRoutines, "goroutines:")
		fmt.Printf("%f%%\n", float64(times[step])/float64(timeout * 1000 * 10))
		nRoutines *= step_size
	}
}

func benchmarkTimeout(n int, nprocs int) int64 {
	runtime.GOMAXPROCS(nprocs)
	var wg sync.WaitGroup
	var timeoutResult int64

	f, err := os.Create("timeout_accuracy-" + strconv.FormatInt(int64(n), 10) + "-procs.pprof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		startTimeout := time.Now()
		c := make(chan int)
		select {
			case <-c:
			case <-time.After(time.Duration(timeout) * time.Millisecond):
		}

		timeoutResult = time.Since(startTimeout).Nanoseconds()
	}(&wg)

	for i := 0; i < n; i++ {
		go func() {
			var j int
			for k := 0; k < iterations; k++ {
				j++
				if (manually_yeild && k%50 == 0) {
					runtime.Gosched()
				}
			}
		}()
	}

	wg.Wait()

	return timeoutResult

}
