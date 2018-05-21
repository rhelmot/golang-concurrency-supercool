package main

import (
	"fmt"
	"sync"
	"golang.org/x/benchmarks/driver"
	"flag"
)

var benchmarkName = *(flag.String("benchmark", "baseline", "which benchmark to run: valid benchmarks: baseline (default), channel-based, buffered-channel-based, shared-memory-based"))
var NPROCS = *(flag.Int("nprocs", 1, "logical CPUs (default 1)"))


func main() {
	benchmarks := make(map[string]func(n uint64))
	benchmarks["baseline"] = baselineCount
	benchmarks["channel-based"] = channelBasedCount
	benchmarks["buffered-channel-based"] = bufferedchannelBasedCount
	benchmarks["shared-memory-based"] = sharedmemBasedCount

	simpleBenchmark(benchmarkName, benchmarks[benchmarkName])
}

func simpleBenchmark(name string, bench func(n uint64)) {
	benchmarkInner := func() driver.Result {
		return driver.Benchmark(bench)
	}
	driver.Main(fmt.Sprintf("%s:%d", name, NPROCS), benchmarkInner)
}

func runBenchmarksWithNprocs(nprocs int) {
	NPROCS = nprocs
	simpleBenchmark("channel-based", channelBasedCount)
	simpleBenchmark("buffered-channel-based", bufferedchannelBasedCount)
	simpleBenchmark("shared-memory-based", sharedmemBasedCount)
}

func baselineCount(n uint64) {
	countWorker(0, int(n), 1)
}

func channelBasedCount(n uint64) {
	results := make(chan int)
	out := 0
	for i := 0; i < NPROCS; i++ {
		go func(i int) {
			results <- countWorker(i, int(n), NPROCS)
		}(i)
	}
	for i := 0; i < NPROCS; i++ {
		out += <-results
	}
}

func bufferedchannelBasedCount(n uint64) {
	results := make(chan int, n)
	out := 0
	for i := 0; i < NPROCS; i++ {
		go func(i int) {
			results <- countWorker(i, int(n), NPROCS)
		}(i)
	}
	for i := 0; i < NPROCS; i++ {
		out += <-results
	}
}

func sharedmemBasedCount(n uint64) {
	results := make([]int, NPROCS)
	wg := sync.WaitGroup{}
	out := 0
	for i := 0; i < NPROCS; i++ {
		wg.Add(1)
		go func(i int) {
			results[i] += countWorker(i, int(n), NPROCS)
			wg.Done()
		}(i)
	}
	wg.Wait()
	for i := 0; i < NPROCS; i++ {
		out += results[i]
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
