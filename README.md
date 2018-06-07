To compile a benchmark:

`go build benchmarks/[benchmark]/[benchmark].go` where `[benchmark]` is data_pipeline, goroutine_startup, parallel_efficiency, or timeout_accuracy.

To run that benchmark:

`./[benchmark]`

The results of the benchmark will be printed on your terminal. A .pprof file will be written to your current directory for every benchmark iteration.