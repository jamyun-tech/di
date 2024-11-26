package di_test

import (
	"testing"

	"github.com/jamyun-tech/di"
	"github.com/stretchr/testify/assert"
)

type (
	QualifiedFoo interface {
		Foo() string
	}

	FirstFooImpl  struct{}
	SecondFooImpl struct{}

	QualifiedBar interface {
		First() string
		Second() string
		Another() string
	}

	QualifiedBarImpl struct {
		first   di.Autowired[QualifiedFoo]
		second  di.Autowired[QualifiedFoo]
		another di.Autowired[QualifiedFoo]
	}
)

func (foo FirstFooImpl) Foo() string {
	return "first foo"
}

func (foo SecondFooImpl) Foo() string {
	return "second foo"
}

func (q QualifiedBarImpl) First() string {
	return q.first().Foo() + " bar"
}

func (q QualifiedBarImpl) Second() string {
	return q.second().Foo() + " bar"
}

func (q QualifiedBarImpl) Another() string {
	return q.another().Foo() + " bar"
}

func TestDIQualifier(t *testing.T) {
	defer di.Release()

	_ = di.Component(&FirstFooImpl{}, new(QualifiedFoo), di.Name("first"))
	_ = di.Component(&SecondFooImpl{}, new(QualifiedFoo), di.Name("second", "another"))
	qualified := di.Component(&QualifiedBarImpl{
		first:   di.Autowire(new(QualifiedFoo), di.Name("first")),
		second:  di.Autowire(new(QualifiedFoo), di.Name("second")),
		another: di.Autowire(new(QualifiedFoo), di.Name("another")),
	}, new(QualifiedBar))

	assert.Equal(t, qualified.First(), "first foo bar")
	assert.Equal(t, qualified.Second(), "second foo bar")
	assert.Equal(t, qualified.Another(), "second foo bar")
}
