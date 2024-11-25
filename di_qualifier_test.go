package di_test

import (
	"github.com/jamyun-tech/di"
	"github.com/stretchr/testify/assert"
	"testing"
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
		FirstFoo   di.Autowired[QualifiedFoo]
		SecondFoo  di.Autowired[QualifiedFoo]
		AnotherFoo di.Autowired[QualifiedFoo]
	}
)

func (foo FirstFooImpl) Foo() string {
	return "first foo"
}

func (foo SecondFooImpl) Foo() string {
	return "second foo"
}

func (q QualifiedBarImpl) First() string {
	return q.FirstFoo().Foo() + " bar"
}

func (q QualifiedBarImpl) Second() string {
	return q.SecondFoo().Foo() + " bar"
}

func (q QualifiedBarImpl) Another() string {
	return q.AnotherFoo().Foo() + " bar"
}

func TestDIQualifier(t *testing.T) {
	defer di.Reset()

	_ = di.Component(&FirstFooImpl{}, new(QualifiedFoo), di.Name("first"))
	_ = di.Component(&SecondFooImpl{}, new(QualifiedFoo), di.Name("second", "another"))
	qualified := di.Component(&QualifiedBarImpl{
		FirstFoo:   di.Autowire(new(QualifiedFoo), di.Name("first")),
		SecondFoo:  di.Autowire(new(QualifiedFoo), di.Name("second")),
		AnotherFoo: di.Autowire(new(QualifiedFoo), di.Name("another")),
	}, new(QualifiedBar))

	assert.Equal(t, qualified.First(), "first foo bar")
	assert.Equal(t, qualified.Second(), "second foo bar")
	assert.Equal(t, qualified.Another(), "second foo bar")
}
