package di_test

import (
	"errors"
	"github.com/jamyun-tech/di"
	"github.com/stretchr/testify/assert"
	"testing"
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
	}
)

func (a SimpleAImpl) DoA() string {
	return "a;"
}

func (b SimpleBImpl) DoB() string {
	return "b;" + b.A().DoA()
}

func TestFailOnBeanDuplication(t *testing.T) {
	defer di.Reset()

	_ = di.Component(new(SimpleA), &SimpleAImpl{})
	assert.Panics(t, func() {
		defer func() {
			errDuplicate := recover().(error)
			assert.NotNil(t, errDuplicate)
			assert.True(t, errors.Is(errDuplicate, di.ErrBeanDuplicate))
			t.Logf("go error: %s", errDuplicate)
			// throw again
			panic(errDuplicate)
		}()

		di.Component(new(SimpleA), &SimpleAImpl{})
	})
}

func TestSimpleAutowire(t *testing.T) {
	defer di.Reset()

	a := di.Component(new(SimpleA), &SimpleAImpl{})
	b := di.Component(new(SimpleB), &SimpleBImpl{
		A: di.Resource(new(SimpleA)),
	})

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
	a := di.Component(new(SimpleA), &SimpleAImpl{})
	b := di.Component(new(CycleB), &CycleBImpl{
		A: di.Resource(new(SimpleA)),
		C: di.Resource(new(CycleC)),
	})
	c := di.Component(new(CycleC), &CycleCImpl{
		A: di.Resource(new(SimpleA)),
		B: di.Resource(new(CycleB)),
	})

	assert.Equal(t, "a;", a.DoA())
	assert.Equal(t, "run:b;a;c;", b.Run())
	assert.Equal(t, "run:c;a;b;", c.Run())
}
