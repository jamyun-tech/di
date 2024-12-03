package di_test

import (
	"testing"

	"github.com/jamyun-tech/di"
	"github.com/stretchr/testify/assert"
)

type (
	SimpleA interface {
		DoA() string
	}

	SimpleAImpl struct{}

	SimpleB interface {
		DoB() string
	}

	SimpleBImpl struct {
		A di.Autowired[SimpleA]
		C di.Autowired[SimpleC]
	}

	SimpleC interface{}
)

func (a SimpleAImpl) DoA() string {
	return "a;"
}

func (b SimpleBImpl) DoB() string {
	return "b;" + b.A().DoA()
}

func TestFailOnBeanDuplication(t *testing.T) {
	defer di.Release()

	_ = di.Component(new(SimpleA), &SimpleAImpl{})
	assertPanicIsError(t, di.ErrBeanDuplicate, func() {
		di.Component(new(SimpleA), &SimpleAImpl{})
	})
}

func TestSimpleAutowire(t *testing.T) {
	defer di.Release()

	a := di.Component(&SimpleAImpl{}, new(SimpleA))
	b := di.Component(&SimpleBImpl{
		A: di.Autowire(new(SimpleA)),
	}, new(SimpleB))

	assert.Equal(t, "a;", di.Autowire(new(SimpleA)).Get().DoA())
	assert.Equal(t, "a;", a.DoA())
	assert.Equal(t, "b;a;", b.DoB())
}

type (
	CycleB interface {
		Run() string
		DoBar() string
	}

	CycleC interface {
		Run() string
		DoC() string
	}

	CycleBImpl struct {
		A di.Autowired[SimpleA]
		C di.Autowired[CycleC]
	}
	CycleCImpl struct {
		A di.Autowired[SimpleA]
		B di.Autowired[CycleB]
	}
)

func (b CycleBImpl) DoBar() string {
	return "b;"
}

func (b CycleBImpl) Run() string {
	return "run:b;" + b.A().DoA() + b.C().DoC()
}

func (c CycleCImpl) DoC() string {
	return "c;"
}

func (c CycleCImpl) Run() string {
	return "run:c;" + c.A().DoA() + c.B().DoBar()
}

func TestCycleAutowire(t *testing.T) {
	defer di.Release()

	a := di.Component(&SimpleAImpl{}, new(SimpleA))
	b := di.Component(&CycleBImpl{
		A: di.Autowire(new(SimpleA)),
		C: di.Autowire(new(CycleC)),
	}, new(CycleB))
	c := di.Component(&CycleCImpl{
		A: di.Autowire(new(SimpleA)),
		B: di.Autowire(new(CycleB)),
	}, new(CycleC))

	assert.Equal(t, "a;", a.DoA())
	assert.Equal(t, "run:b;a;c;", b.Run())
	assert.Equal(t, "run:c;a;b;", c.Run())
}

func TestDefaultCycleAutowire(t *testing.T) {
	defer di.Release()

	a := di.Component(&SimpleAImpl{})
	b := di.Component(&CycleBImpl{
		A: di.Autowire(new(SimpleA)),
		C: di.Autowire(new(CycleC)),
	})
	c := di.Component(&CycleCImpl{
		A: di.Autowire(new(SimpleA)),
		B: di.Autowire(new(CycleB)),
	})

	assert.Equal(t, "a;", a.DoA())
	assert.Equal(t, "run:b;a;c;", b.Run())
	assert.Equal(t, "run:c;a;b;", c.Run())
}

type (
	StructA struct{}
	StructB struct {
		A di.Autowired[*StructA]
	}
)

func TestStructTypeReference(t *testing.T) {
	defer di.Release()

	di.Component(&StructA{}, new(*StructA))
	di.Component(&StructB{
		A: di.Autowire(new(*StructA)),
	}, new(*StructB))

	assert.NotPanics(t, di.Validate)
}

func TestDefaultStructTypeReference(t *testing.T) {
	defer di.Release()

	di.Component(&StructA{})
	di.Component(&StructB{
		A: di.Autowire(new(*StructA)),
	})

	assert.NotPanics(t, di.Validate)
}

func TestCannotRegisterNilBean(t *testing.T) {
	di.Release()

	assertPanicIsError(t, di.ErrBeanNil, func() {
		var nilBean *SimpleAImpl = nil
		di.Component(nilBean, new(SimpleA))
	})
	assertPanicIsError(t, di.ErrBeanNil, func() {
		di.Component((*SimpleAImpl)(nil), new(SimpleA))
	})
}

func assertPanicIsError(t *testing.T, target error, panicFunc func()) {
	assert.Panics(t, func() {
		defer func() {
			err := recover().(error)
			assert.ErrorIs(t, err, target)
			t.Logf("got expected panic: %s", err)
			panic(err)
		}()
		panicFunc()
	})
}

func TestFailOnValidate(t *testing.T) {
	defer di.Release()

	assert.NotPanics(t, func() {
		di.Component(&SimpleAImpl{}, new(SimpleA))
		di.Component(&SimpleBImpl{
			A: di.Autowire(new(SimpleA)),
			C: di.Autowire(new(SimpleC)),
		}, new(SimpleB))
	})

	assertPanicIsError(t, di.ErrBeanNotFound, func() {
		di.Validate()
	})
}
