# Installation
- Install vips binary



## Build
to build a image
`DOCKER_BUILDKIT=1 docker build -t pdf2img --target prod . `



## Benchmark
```go test -benchmem -run=^$ -bench ^BenchmarkConvertPDFToImage$ github.com/felixgao/pdf_to_png  &>> benchmark.log```

```
-bench=. - Tells go test to run benchmarks and tests within the project's _test.go files. . is a regular expression that tells go test to match with everything.

./pdf_to_png - Tells go test the location of the _test.go files with the benchmarks and tests to run.

-run="^$" - Tells go test to run tests with names that satisfy the regular expression ^$. Since none of the tests' names begin with $, go test will not run any tests and will only run benchmarks.  
```


```bash
go test -benchmem -bench="^BenchmarkConvertPDFToImage$" -run="^$" -cpu=1,2,4,8 -benchtime=10s -cpuprofile=cpu.out -memprofile=mem.out -trace=trace.out | tee bench.txt

go tool pprof -http :8080 cpu.out
go tool pprof -http :8081 mem.out
go tool trace trace.out


# go install golang.org/x/perf/cmd/benchstat
benchstat bench.txt
rm cpu.out mem.out trace.out *.test
```


