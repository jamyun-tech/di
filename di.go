package di

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"sync"
)

var (
	ErrBeanNotFound  = errors.New("bean not found")
	ErrBeanDuplicate = errors.New("bean already exists")
	global           = &AppContext{}
)

type (
	Autowired[T any] func() T

	AppContext struct {
		register sync.Map // bean cache: type -> bean
		alias    sync.Map // qualified cache: string -> bean
	}

	Qualifier interface {
		Qualified() []string
	}

	Foo interface {
		Bar() Bar
	}

	Bar interface {
		Foo() Foo
	}
)

func Reset() {
	global.register.Clear()
	global.alias.Clear()
}

func TypeOf(def any) reflect.Type {
	return reflect.TypeOf(def).Elem()
}

func Component[T any](def any, bean T) T {
	global.Component(def, bean)
	return bean
}

func (ac *AppContext) Component(def, bean any) any {
	return ac.TComponent(TypeOf(def), bean)
}

func (ac *AppContext) TComponent(def reflect.Type, bean any) any {
	if _, exist := ac.register.LoadOrStore(def, bean); exist {
		panic(errors.Wrapf(ErrBeanDuplicate, "fail when register bean[%s]", def))
	}
	return bean
}

func Resource[T any](def *T) Autowired[T] {
	wired := global.Resource(def)
	return sync.OnceValue(func() T {
		bean := wired()
		return bean.(T)
	})
}

func (ac *AppContext) Resource(def any) Autowired[any] {
	return ac.TResource(TypeOf(def))
}

func (ac *AppContext) TResource(def reflect.Type) Autowired[any] {
	bean, ok := ac.register.Load(def)
	if ok {
		return func() any {
			return bean
		}
	}
	return ac.lazyLoad(def)
}

func (ac *AppContext) lazyLoad(def reflect.Type) func() any {
	return sync.OnceValue(func() any {
		bean, ok := ac.register.Load(def)
		if ok {
			return bean
		}
		panic(errors.Wrap(ErrBeanNotFound, fmt.Sprintf(", type: %s", def)))
	})
}
