package utils

import (
	"reflect"
	"strings"
)

func TrimStruct(obj any, skip ...string) {
	v := reflect.ValueOf(obj)

	if v.Kind() != reflect.Ptr {
		return
	}

	v = v.Elem()

	skipSet := make(map[string]bool)
	for _, s := range skip {
		skipSet[s] = true
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		value := v.Field(i)

		if value.Kind() == reflect.String && value.CanSet() && !skipSet[field.Name] {
			value.SetString(strings.TrimSpace(value.String()))
		}
	}
}
