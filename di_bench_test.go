package di_test

import (
	"fmt"
	"github.com/jamyun-tech/di"
	"testing"
	"time"
)

type (
	PlainA interface {
		Run() struct{}
	}
	PlainB interface {
		Run() struct{}
	}
	PlainAImpl struct{}
	PlainBImpl struct {
		A PlainA
	}

	DIA interface {
		Run() struct{}
	}
	DIB interface {
		Run() struct{}
	}
	DIAImpl struct{}
	DIBImpl struct {
		A di.Autowired[DIA]
	}
)

func (p PlainAImpl) Run() struct{} {
	return struct{}{}
}

func (p PlainBImpl) Run() struct{} {
	return p.A.Run()
}

func (d DIAImpl) Run() struct{} {
	return struct{}{}
}

func (d DIBImpl) Run() struct{} {
	return d.A().Run()
}

func BenchmarkPlainStruct(b *testing.B) {
	defer di.Reset()

	pa := &PlainAImpl{}
	pb := &PlainBImpl{pa}
	for i := 0; i < b.N; i++ {
		pb.Run()
	}
}

func BenchmarkDIStruct(b *testing.B) {
	defer di.Reset()

	di.Component(new(DIA), &DIAImpl{})
	db := di.Component(new(DIB), &DIBImpl{
		A: di.Resource(new(DIA)),
	})
	for i := 0; i < b.N; i++ {
		db.Run()
	}
}

type (
	FmtA interface {
		Run() string
	}
	FmtB interface {
		Run() string
	}
	FmtAImpl struct{}
	FmtBImpl struct {
		A FmtA
	}

	DiFmtA interface {
		Run() string
	}
	DiFmtB interface {
		Run() string
	}
	DiFmtAImpl struct{}
	DiFmtBImpl struct {
		A di.Autowired[DiFmtA]
	}
)

func (d FmtAImpl) Run() string {
	return fmt.Sprintf("time: %d", time.Now().UnixNano())
}

func (d FmtBImpl) Run() string {
	return fmt.Sprintf("from b: %s", d.A.Run())
}

func (d DiFmtAImpl) Run() string {
	return fmt.Sprintf("time: %d", time.Now().UnixNano())
}

func (d DiFmtBImpl) Run() string {
	return fmt.Sprintf("from b: %s", d.A().Run())
}

func BenchmarkPlainFmt(b *testing.B) {
	fa := &FmtAImpl{}
	fb := &FmtBImpl{fa}

	for i := 0; i < b.N; i++ {
		fb.Run()
	}
}

func BenchmarkDiFmt(b *testing.B) {
	defer di.Reset()

	di.Component(new(DiFmtA), &DiFmtAImpl{})
	db := di.Component(new(DiFmtB), &DiFmtBImpl{
		A: di.Resource(new(DiFmtA)),
	})
	for i := 0; i < b.N; i++ {
		db.Run()
	}
}
