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
		bar di.Autowired[Bar]
	}
	BarImpl struct {
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
	foo := di.Component(new(Foo), &FooImpl{
		bar: di.Resource(new(Bar)),
	})
	bar := di.Component(new(Bar), &BarImpl{
		foo: di.Resource(new(Foo)),
	})

	fmt.Println(foo.Run())
	fmt.Println(bar.Run())
	if foo.Run() != "foo;bar;" {
		panic("oops")
	}
}
