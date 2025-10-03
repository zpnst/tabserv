package ioc

import (
	"fmt"
	"reflect"
	"sync"
)

type Container struct {
	providers     map[reflect.Type]func(c *Container) any
	providersLock sync.RWMutex
}

func NewContainer() *Container {
	return &Container{
		providers: make(map[reflect.Type]func(*Container) any),
	}
}

func Register[T any](c *Container, provider func(c *Container) T) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	c.providersLock.Lock()
	c.providers[t] = func(c *Container) any { return provider(c) }
	c.providersLock.Unlock()
}

func RegisterSingleton[T any](c *Container, provider func(c *Container) T) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	var once sync.Once
	var inst any
	c.providersLock.Lock()
	c.providers[t] = func(c *Container) any {
		once.Do(func() {
			inst = provider(c)
		})
		return inst
	}
	c.providersLock.Unlock()
}

func Resolve[T any](c *Container) T {
	t := reflect.TypeOf((*T)(nil)).Elem()
	c.providersLock.RLock()
	prov, ok := c.providers[t]
	c.providersLock.RUnlock()
	if !ok {
		panic(fmt.Errorf("no provider for %v", t))
	}
	return prov(c).(T)
}
