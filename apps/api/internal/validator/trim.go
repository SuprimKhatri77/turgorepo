package validator

import (
	"reflect"
	"strings"
)

// trimStrings trims surrounding whitespace from the top-level string fields of
// the struct obj points to. Fields tagged `trim:"-"` are left alone, for values
// where spaces are meaningful, like passwords.
func trimStrings(obj any) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return
	}
	v = v.Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		value := v.Field(i)
		if value.Kind() == reflect.String && value.CanSet() && field.Tag.Get("trim") != "-" {
			value.SetString(strings.TrimSpace(value.String()))
		}
	}
}
