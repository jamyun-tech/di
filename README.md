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
result shows average overhead per ops is 1.5~1.7ns. In real world with real work load, the overhead can be ignored.

```shell
goos: darwin
goarch: arm64
pkg: github.com/jamyun-tech/di
cpu: Apple M2 Max
BenchmarkPlainStruct
BenchmarkPlainStruct-12    	592450573	         2.033 ns/op
BenchmarkDIStruct
BenchmarkDIStruct-12       	340340929	         3.511 ns/op
BenchmarkPlainFmt
BenchmarkPlainFmt-12       	 7590954	       162.3 ns/op
BenchmarkDiFmt
BenchmarkDiFmt-12          	 7383570	       160.9 ns/op
PASS
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
	Foo interface {
		DoFoo() string
		Run() string
	}

	Bar interface {
		DoBar() string
		Run() string
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
		bar: di.Resource(new(Bar)),
	}, new(Foo))
	bar := di.Component(&BarImpl{
		foo: di.Resource(new(Foo)),
	}, new(Bar))

	fmt.Println(foo.Run())
	fmt.Println(bar.Run())
	if foo.Run() != "foo;bar;" {
		panic("oops")
	}
}
```

More example: [Example](https://github.com/jamyun-tech/di/tree/main/example)
