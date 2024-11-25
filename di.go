package di

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

var (
	ErrBeanNil        = errors.New("bean cannot be nil")
	ErrBeanDefinition = errors.New("bean definition error")
	ErrBeanNotFound   = errors.New("bean not found")
	ErrBeanDuplicate  = errors.New("bean already exists")
	global            = &AppContext{}
)

type (
	Autowired[T any] func() T
	Describe         func(d *Definition)

	Definition struct {
		Qualifier []string
		BeanType  reflect.Type
		Bean      any
	}

	AppContext struct {
		register   sync.Map // bean cache: type -> bean
		definition sync.Map // qualified cache: type -> config
		validator  []func()
	}

	Qualifier interface {
		Qualified() []string
	}
)

func Reset() {
	global.register.Clear()
	global.definition.Clear()
}

func Validate() {
	global.Validate()
}

func (ac *AppContext) Validate() {
	if len(ac.validator) == 0 {
		return
	}
	for _, v := range ac.validator {
		// should fail now if any bean initialize fail
		v()
	}
}

func TypeOf(def any) reflect.Type {
	return reflect.TypeOf(def).Elem()
}

func Component[T any](bean T, beanType any, describes ...Describe) T {
	global.Component(bean, beanType, describes...)
	return bean
}

func (ac *AppContext) Component(bean, beanType any, describes ...Describe) any {
	return ac.TComponent(bean, TypeOf(beanType), describes...)
}

func (ac *AppContext) TComponent(bean any, beanType any, describes ...Describe) any {
	if bean == nil || reflect.ValueOf(bean).IsNil() {
		panic(fmt.Errorf("%w: [%s] cannot be nil", ErrBeanNil, reflect.TypeOf(bean).Elem()))
	}

	var definition = Definition{Bean: bean}
	if realType, ok := beanType.(reflect.Type); ok {
		definition.BeanType = realType
	} else if describe, ok := beanType.(Describe); ok {
		describes = append([]Describe{describe}, describes...)
	}
	for _, describe := range describes {
		describe(&definition)
	}
	if definition.BeanType == nil {
		panic(fmt.Errorf("%w: [%s] type unknown", ErrBeanDefinition, reflect.TypeOf(beanType).Elem()))
	}

	if _, exist := ac.register.LoadOrStore(definition.BeanType, bean); exist {
		panic(fmt.Errorf("%w, [%s] duplicate defifnition", ErrBeanDuplicate, beanType))
	} else {
		ac.definition.Store(definition.BeanType, definition)
	}
	return bean
}

func AutowireAll[T any](def *T, cfg ...Describe) Autowired[[]T] {
	// TODO
	return nil
}

func Autowire[T any](beanType *T, describes ...Describe) Autowired[T] {
	wired := global.Resource(beanType, describes...)
	return sync.OnceValue(func() T {
		return wired().(T)
	})
}

func (ac *AppContext) Resource(beanType any, describes ...Describe) Autowired[any] {
	wired := ac.TResource(TypeOf(beanType), describes...)
	ac.validator = append(ac.validator, lazyValidate(wired))
	return wired
}

func (ac *AppContext) TResource(def reflect.Type, describes ...Describe) Autowired[any] {
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
		panic(fmt.Errorf("%w, type: %s", ErrBeanNotFound, def))
	})
}

func lazyValidate(beanWire Autowired[any]) func() {
	return func() {
		beanWire()
	}
}

func Name(name ...string) Describe {
	return func(config *Definition) {
		if len(name) == 0 {
			config.Qualifier = []string{"default"}
		} else {
			config.Qualifier = name
		}
	}
}
