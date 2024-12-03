package di_test

import (
	"github.com/jamyun-tech/di"
	"github.com/stretchr/testify/assert"
	"testing"
)

type (
	BatchFoo interface {
		DoBatchFoo() string
	}

	FirstBatchFoo  struct{}
	SecondBatchFoo struct{}
	ThirdBatchFoo  struct{}
)

func (FirstBatchFoo) DoBatchFoo() string {
	return "first"
}

func (SecondBatchFoo) DoBatchFoo() string {
	return "second"
}

func (ThirdBatchFoo) DoBatchFoo() string {
	return "third"
}

func TestAutowireAll(t *testing.T) {
	_ = di.Component(&FirstBatchFoo{}, new(BatchFoo), di.Name("first"))
	_ = di.Component(&SecondBatchFoo{}, new(BatchFoo), di.Name("second"))
	_ = di.Component(&ThirdBatchFoo{}, new(BatchFoo), di.Name("third"))

	beans := di.AutowireAll(new(BatchFoo))
	assert.Len(t, beans(), 3)
}
