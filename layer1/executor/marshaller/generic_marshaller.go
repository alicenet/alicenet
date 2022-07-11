package objects

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

var (
	ErrUnknownName = errors.New("unknown type name")
	ErrUnknownType = errors.New("unknown instance type")
)

type TypeRegistry struct {
	sync.RWMutex
	sync.Once
	a map[reflect.Type]string
	b map[string]reflect.Type
}

type InstanceWrapper struct {
	NameType    string
	RawInstance []byte
}

func (registry *TypeRegistry) RegisterInstanceType(t interface{}) {
	registry.Lock()
	defer registry.Unlock()

	registry.Do(func() {
		registry.a = make(map[reflect.Type]string)
		registry.b = make(map[string]reflect.Type)
	})

	name, tipe := GetNameType(t)

	registry.a[tipe] = name
	registry.b[name] = tipe
}

func (registry *TypeRegistry) lookupName(tipe reflect.Type) (string, bool) {
	registry.RLock()
	defer registry.RUnlock()

	present, name := registry.a[tipe]

	return present, name
}

func (registry *TypeRegistry) lookupType(name string) (reflect.Type, bool) {
	registry.RLock()
	defer registry.RUnlock()

	present, tipe := registry.b[name]

	return present, tipe
}

func (registry *TypeRegistry) WrapInstance(t interface{}) (*InstanceWrapper, error) {

	tipe := reflect.TypeOf(t)
	if tipe.Kind() == reflect.Ptr {
		tipe = tipe.Elem()
	}

	name, present := registry.lookupName(tipe)
	if !present {
		panic(fmt.Errorf("unable to wrapInstance: %v", tipe))
	}

	raw, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	return &InstanceWrapper{NameType: name, RawInstance: raw}, nil
}

func (registry *TypeRegistry) UnwrapInstance(wrapper *InstanceWrapper) (interface{}, error) {

	tipe, present := registry.lookupType(wrapper.NameType)
	if !present {
		return nil, ErrUnknownName
	}
	val := reflect.New(tipe)

	err := json.Unmarshal(wrapper.RawInstance, val.Interface())
	if err != nil {
		return nil, err
	}

	return val.Interface(), nil
}

func GetNameType(t interface{}) (string, reflect.Type) {
	tipe := reflect.TypeOf(t)
	if tipe.Kind() == reflect.Ptr {
		tipe = tipe.Elem()
	}

	return tipe.String(), tipe
}
