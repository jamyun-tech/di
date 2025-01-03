package di

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

var (
	ErrBeanNil           = errors.New("bean cannot be nil")
	ErrBeanTypeAmbiguous = errors.New("bean type ambiguous")
	ErrBeanDefinition    = errors.New("bean definition error")
	ErrBeanNotFound      = errors.New("bean not found")
	ErrBeanDuplicate     = errors.New("bean already exists")
	global               = &AppContext{register: make(map[*Definition]struct{})}
)

type (
	Autowired[T any]    func() T
	AutowiredAll[T any] func() []T
	Describe            func(d *Definition)

	// Definition defines a bean by it's type by BeanType
	// and with it's alias by Qualifier, Qualifier is a comma
	// seperate string.
	Definition struct {
		Qualifier []string
		BeanType  reflect.Type
		Bean      any
		Wildcard  bool
	}

	AppContext struct {
		register  map[*Definition]struct{} // bean definition set: {definition, ...}
		validator []func()
		rw        sync.RWMutex
	}

	Qualifier interface {
		Qualified() []string
	}
)

func (a Autowired[T]) Get() T {
	return a()
}

func Release() {
	global.Release()
}

func Validate() {
	global.Validate()
}

func (ac *AppContext) Validate() {
	ac.rw.Lock()
	if len(ac.validator) == 0 {
		ac.rw.Unlock()
		return
	}
	snapshot := make([]func(), len(ac.validator))
	copy(snapshot, ac.validator)
	ac.rw.Unlock()

	// should fail now if any bean initialize fail
	for _, v := range snapshot {
		v()
	}
}

func (ac *AppContext) Release() {
	ac.rw.Lock()
	defer ac.rw.Unlock()

	ac.register = make(map[*Definition]struct{})
	ac.validator = nil
}

func TypeOf(def any) *Definition {
	if reflectType, ok := def.(reflect.Type); ok {
		return &Definition{BeanType: reflectType, Wildcard: true}
	}
	return &Definition{BeanType: reflect.TypeOf(def).Elem(), Wildcard: true}
}

func Component[T any](bean T, describes ...any) T {
	global.Component(bean, describes...)
	return bean
}

func (ac *AppContext) Component(bean any, describes ...any) any {
	var (
		beanType any = reflect.TypeOf(bean)
		beanDef  any = nil
		desc     []Describe
	)
	if len(describes) > 0 {
		for _, e := range describes {
			if d, ok := e.(Describe); ok {
				// collect all bean describe
				desc = append(desc, d)
			} else {
				// should have only one bean type definition
				if beanDef != nil {
					panic(fmt.Errorf("%w: bean[%s] has multipul type %s, %s",
						ErrBeanTypeAmbiguous, reflect.TypeOf(bean), reflect.TypeOf(beanDef), reflect.TypeOf(e)))
				}
				beanDef = e
			}
		}
	}
	if beanDef != nil {
		beanType = beanDef
	}

	return ac.TComponent(bean, TypeOf(beanType), desc...)
}

func (ac *AppContext) TComponent(bean any, beanType any, describes ...Describe) any {
	ac.rw.Lock()
	defer ac.rw.Unlock()

	if bean == nil || reflect.ValueOf(bean).IsNil() {
		panic(fmt.Errorf("%w: [%s] cannot be nil", ErrBeanNil, reflect.TypeOf(bean).Elem()))
	}

	definition := &Definition{Bean: bean}
	if def, ok := beanType.(*Definition); ok {
		definition.BeanType = def.BeanType
	}
	// todo: unused code
	// else if describe, ok := beanType.(Describe); ok {
	// 	describes = append([]Describe{describe}, describes...)
	// }

	definition.Apply(describes)

	if definition.BeanType == nil {
		panic(fmt.Errorf("%w: [%s] type unknown", ErrBeanDefinition, reflect.TypeOf(beanType).Elem()))
	}

	if _, exist := ac.loadOne(definition); exist {
		panic(fmt.Errorf("%w, [%s] duplicate defifnition", ErrBeanDuplicate, definition.BeanType))
	} else {
		ac.register[definition] = struct{}{}
	}

	return bean
}

func AutowireAll[T any](beanType *T, describes ...Describe) AutowiredAll[T] {
	var (
		result []T
		once   sync.Once
		wired  = global.AutowireAll(beanType, describes...)
	)
	g := func() {
		beans := wired()
		if len(beans) == 0 {
			result = nil
			return
		}
		result = make([]T, len(beans))
		for i := 0; i < len(beans); i++ {
			result[i] = beans[i].(T)
		}
		wired = nil
	}
	return func() []T {
		once.Do(g)
		return result
	}
}

func (ac *AppContext) AutowireAll(beanType any, describes ...Describe) AutowiredAll[any] {
	var (
		result []any
		once   sync.Once
		wired  = ac.FindAll(TypeOf(beanType), describes...)
	)
	g := func() {
		result = wired()
		wired = nil
	}
	return func() []any {
		once.Do(g)
		return result
	}
}

func (ac *AppContext) FindAll(def *Definition, describes ...Describe) AutowiredAll[any] {
	def.Apply(describes)
	return func() []any {
		ac.rw.RLock()
		defer ac.rw.RUnlock()

		bean, ok := ac.loadAll(def)
		if ok {
			return bean
		}
		panic(fmt.Errorf("%w, type: %s", ErrBeanNotFound, def.BeanType))
	}
}

func (ac *AppContext) loadAll(d *Definition) (beans []any, exist bool) {
	for def := range ac.register {
		if def.Match(d) {
			beans = append(beans, def.Bean)
			exist = true
		}
	}
	return
}

func Autowire[T any](beanType *T, describes ...Describe) Autowired[T] {
	var (
		result T
		once   sync.Once
		wired  = global.Autowire(beanType, describes...)
	)
	g := func() {
		result = wired().(T)
		wired = nil
	}
	return func() T {
		once.Do(g)
		return result
	}
}

func (ac *AppContext) Autowire(beanType any, describes ...Describe) Autowired[any] {
	var (
		result any
		once   sync.Once
		wired  = ac.Find(TypeOf(beanType), describes...)
	)
	g := func() {
		result = wired()
		ac.validator = append(ac.validator, lazyValidate(wired))
		wired = nil
	}
	return func() any {
		once.Do(g)
		return result
	}
}

func (ac *AppContext) Find(def *Definition, describes ...Describe) Autowired[any] {
	ac.rw.RLock()
	defer ac.rw.RUnlock()

	def.Apply(describes)

	bean, ok := ac.loadOne(def)
	if ok {
		return func() any {
			return bean
		}
	}
	// lazy load
	return func() any {
		ac.rw.RLock()
		defer ac.rw.RUnlock()

		bean, ok := ac.loadOne(def)
		if ok {
			return bean
		}
		panic(fmt.Errorf("%w, type: %s", ErrBeanNotFound, def.BeanType))
	}
}

func (ac *AppContext) loadOne(d *Definition) (bean any, exist bool) {
	for def := range ac.register {
		if def.Match(d) {
			return def.Bean, true
		}
	}
	return nil, false
}

func (d *Definition) Apply(describes []Describe) {
	if len(describes) == 0 {
		return
	}

	for _, describe := range describes {
		describe(d)
	}
}

func (d *Definition) Match(o *Definition) bool {
	if d.BeanType == o.BeanType || (o.Wildcard && d.BeanType.AssignableTo(o.BeanType)) {
		if len(d.Qualifier) == 0 || len(d.Qualifier) == 1 && d.Qualifier[0] == "default" {
			return true
		}
		if len(o.Qualifier) > 0 && len(d.Qualifier) > 0 {
			for _, name := range d.Qualifier {
				for _, out := range o.Qualifier {
					if name == out {
						return true
					}
				}
			}
			return false
		}
		return true
	}
	return false
}

func lazyValidate(beanWire Autowired[any]) func() {
	return func() {
		beanWire()
	}
}

func lazyValidateAll(beanWire AutowiredAll[any]) func() {
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
func DisableWildcard() Describe {
	return func(config *Definition) {
		config.Wildcard = false
	}
}
