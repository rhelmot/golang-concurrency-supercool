package main

import (
    "os"
	"fmt"
	"flag"
    "time"
    "runtime"
)

var max_goroutines_flag = flag.Int("max_goroutines", runtime.NumCPU() * 2, "maximum number of parallel goroutines to use (default: number of CPU cores * 2)")
var task_difficulty_flag = flag.Int("task_difficulty", 10000, "difficulty of individual tasks which will be parallelized across goroutines (default 10000)")
var iterations_flag = flag.Int("iterations", 1000, "number of times to run task parallelized across goroutines")

func main() {
	flag.Parse()
    max_goroutines := *max_goroutines_flag
    task_difficulty := *task_difficulty_flag
    iterations := *iterations_flag

    timesCycled := make(map[int]time.Duration)
    timesLive := make(map[int]time.Duration)

    fmt.Println("Computing canonical result")
    canonResult := countWorker(0, task_difficulty, 1)
    var start time.Time

    for nworkers := 1; nworkers <= max_goroutines; nworkers++ {
        fmt.Fprintf(os.Stderr, "Testing cycled nworkers=%d\n", nworkers)
        runtime.GC()
        start = time.Now()
        runCycled(task_difficulty, nworkers, iterations, canonResult)
        timesCycled[nworkers] = time.Since(start) / time.Duration(iterations)

        fmt.Fprintf(os.Stderr, "Testing live nworkers=%d\n", nworkers)
        runtime.GC()
        start = time.Now()
        runLive(task_difficulty, nworkers, iterations, canonResult)
        timesLive[nworkers] = time.Since(start) / time.Duration(iterations)
    }

    fmt.Fprintln(os.Stderr, "n   cycled        live")
    for nworkers := 1; nworkers <= max_goroutines; nworkers++ {
        fmt.Printf("%02d  %-13v %-13v\n",
            nworkers,
            timesCycled[nworkers],
            timesLive[nworkers])
    }
}

func runLive(count, nworkers, trials, canonResult int) {
    starts := make(chan int)
    results := make(chan int)
    quit := make(chan int, 1)

    for i := 0; i < nworkers; i++ {
        go liveWorker(starts, results, quit, count, nworkers)
    }

    for l := 0; l < trials; l++ {
        for i := 0; i < nworkers; i++ {
            starts <- i
        }
        result := 0
        for i := 0; i < nworkers; i++ {
            result += <-results
        }
        if result != canonResult {
            panic("Bad result")
        }
    }

    for i := 0; i < nworkers; i++ {
        quit <- 1
    }
}

func runCycled(count, nworkers, trials, canonResult int) {
    for l := 0; l < trials; l++ {
        results := make(chan int)
        result := 0
        for i := 0; i < nworkers; i++ {
            go func(i int) {
                results <- countWorker(i, count, nworkers)
            }(i)
        }
        for i := 0; i < nworkers; i++ {
            result += <-results
        }
        if result != canonResult {
            panic("Bad result")
        }
    }
}

func liveWorker(starts, results, quit chan int, end, stride int) {
    for {
        select {
        case <-quit:
            return
        case start := <-starts:
            results <- countWorker(start, end, stride)
        }
    }
}

func countWorker(start, end, stride int) int {
	out := 0
	for ; start < end; start += stride {
		if isPrime(start) {
			out += 1
		}
	}
	return out
}

func isPrime(x int) bool {
	if x <= 1 {
		return false;
	}
	for y := 2; y*y <= x; y++ {
		if x % y == 0 {
			return false
		}
	}
	return true
}
