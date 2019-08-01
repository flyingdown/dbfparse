package dbfparse

import (
	"fmt"
	"reflect"
)

var typeRegistry = make(map[string]reflect.Type)

func Register(elem interface{}) {
	t := reflect.TypeOf(elem).Elem()
	typeRegistry[t.Name()] = t
}

func NewObject(name string) (interface{}, error) {
	if typ, ok := typeRegistry[name]; ok {
		return reflect.New(typ).Interface(), nil
	} else {
		return nil, fmt.Errorf("not found %s struct", name)
	}
}
