# di: Single file dependency inject manager

Since Golang ship new feature like [generic](https://go.dev/doc/tutorial/generics), Finally we can have a usable DI library.
`di` looks like spring context (but with functional style), provide resource management at start time with reflect and 
auto type cast, and yes `di` will **fail fast** to avoid runtime exception.

# Getting start

When using `di`, a `Autowired` value is a lazy supplier function.

```go
// wrap value with a supplier function
type Autowired[T any] func() T

// comparing to a plain Bean
var Foo IFaceFoo
// use supplier style value declaration
var Bar Autowired[IFaceBar]
```

Change your plain bean definition to a autowired bean definition:

```go
// a plain bean example
type PlainFoo struct {
	bar Bar
}

func (foo PlainFoo) Run() {
	foo.bar.DoSth()
}

// a autowired bean example
type DiFoo struct {
	bar Autowired[Bar]
}

func (foo DiFoo) Run() {
	foo.bar().DoSth()
}
```

# Performance Overhead

`di` wrap value with `sync.OnceValue` and assemble bean using a map as a bean registry, the performance overhead is
acceptable.

We'll have the following [benchmark](https://github.com/jamyun-tech/di/blob/main/di_bench_test.go) running on my laptop, the
result shows average overhead per ops is 0.5~0.7ns. In real world with real work load, the overhead can be ignored.

```shell
go test -benchtime=10s -bench .

goos: darwin
goarch: arm64
pkg: github.com/jamyun-tech/di
cpu: Apple M2 Max
BenchmarkPlainStruct-12    	1000000000	         2.169 ns/op
BenchmarkDIStruct-12       	1000000000	         2.662 ns/op
BenchmarkPlainFmt-12       	77954832	       152.6 ns/op
BenchmarkDiFmt-12          	77462708	       153.3 ns/op
PASS
ok  	github.com/jamyun-tech/di	29.403s
```

# Examples

And of course, `di` can handle **cycle**-referenced bean with lazy loader.

```go
package main

import (
	"fmt"
	"github.com/jamyun-tech/di"
)

type (
	Runner interface {
		Run() string
	}

	Foo interface {
		Runner
		DoFoo() string
	}

	Bar interface {
		Runner
		DoBar() string
	}

	Tar interface {
		Bar
	}

	FooImpl struct {
		// cycle reference here!
		bar di.Autowired[Bar]
	}
	BarImpl struct {
		// cycle reference here!
		foo di.Autowired[Foo]
	}
)

func (a FooImpl) DoFoo() string {
	return "foo;"
}

func (a FooImpl) Run() string {
	return "foo;" + a.bar().DoBar()
}

func (b BarImpl) DoBar() string {
	return "bar;"
}

func (b BarImpl) Run() string {
	return "bar;" + b.foo().DoFoo()
}

func main() {
	foo := di.Component(&FooImpl{
		bar: di.Autowire(new(Bar)),
	})
	bar := di.Component(&BarImpl{
		foo: di.Autowire(new(Foo)),
	})

	fmt.Println(foo.Run())
	fmt.Println(bar.Run())
	if foo.Run() != "foo;bar;" {
		panic("oops")
	}
}
```

More example: [Example](https://github.com/jamyun-tech/di/tree/main/example)
