package di_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jamyun-tech/di"
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
	DIBEagerImpl struct {
		A DIA
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

func (d DIBEagerImpl) Run() struct{} {
	return d.A.Run()
}

func BenchmarkPlainStruct(b *testing.B) {
	defer di.Release()

	pa := &PlainAImpl{}
	pb := &PlainBImpl{pa}
	for i := 0; i < b.N; i++ {
		pb.Run()
	}
}

func BenchmarkDIStruct(b *testing.B) {
	defer di.Release()

	di.Component(&DIAImpl{}, new(DIA))
	db := di.Component(&DIBImpl{
		A: di.Autowire(new(DIA)),
	}, new(DIB))
	for i := 0; i < b.N; i++ {
		db.Run()
	}
}

func BenchmarkDIEagerStruct(b *testing.B) {
	defer di.Release()

	di.Component(&DIAImpl{}, new(DIA))
	db := di.Component(&DIBEagerImpl{
		A: di.Autowire(new(DIA)).Get(),
	}, new(DIB))
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
	DiFmtBEagerImpl struct {
		A DiFmtA
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

func (d DiFmtBEagerImpl) Run() string {
	return fmt.Sprintf("from b: %s", d.A.Run())
}

func BenchmarkPlainFmt(b *testing.B) {
	fa := &FmtAImpl{}
	fb := &FmtBImpl{fa}

	for i := 0; i < b.N; i++ {
		fb.Run()
	}
}

func BenchmarkDiFmt(b *testing.B) {
	defer di.Release()

	di.Component(&DiFmtAImpl{}, new(DiFmtA))
	db := di.Component(&DiFmtBImpl{
		A: di.Autowire(new(DiFmtA)),
	}, new(DiFmtB))
	for i := 0; i < b.N; i++ {
		db.Run()
	}
}

func BenchmarkDiEagerFmt(b *testing.B) {
	defer di.Release()

	di.Component(&DiFmtAImpl{}, new(DiFmtA))
	db := di.Component(&DiFmtBEagerImpl{
		A: di.Autowire(new(DiFmtA)).Get(),
	}, new(DiFmtB))
	for i := 0; i < b.N; i++ {
		db.Run()
	}
}
