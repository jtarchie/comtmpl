# comtmpl

Compile `html/template` to go code.

## Why?

I wanted to be able to convert some of my templates directly to Go code. This
was part exercise and part, I know string writing can be faster for my one API
endpoint.

## Benchmark

From examples,

```
pkg: github.com/jtarchie/comtmpl/examples
cpu: Apple M4
BenchmarkStandardTemplate-10    	 1080583	      1104 ns/op	     736 B/op	      31 allocs/op
BenchmarkCustomTemplate-10      	12625918	        96.39 ns/op	       0 B/op	       0 allocs/op
```

## Test

```
task
```
