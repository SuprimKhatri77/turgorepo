package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
)

func Parse(err error, obj any) []response.FieldError {
	var ve validator.ValidationErrors
	var typeErr *json.UnmarshalTypeError

	if errors.As(err, &typeErr) && typeErr.Field != "" {
		return []response.FieldError{typeError(typeErr, obj)}
	}
	if !errors.As(err, &ve) {
		return []response.FieldError{
			{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		}
	}

	var errs []response.FieldError

	for _, fe := range ve {
		field := getStructField(obj, fe)
		jsonField := jsonName(field, fe)
		label := fieldLabel(field, jsonField)

		errs = append(errs, response.FieldError{
			Code:    mapTagToCode(fe.Tag()),
			Field:   jsonField,
			Message: messageFor(fe, field, label),
		})
	}

	return errs
}

// typeError names the field whose JSON value had the wrong type, such as a
// number sent for a text field, so the client can show the error there.
func typeError(err *json.UnmarshalTypeError, obj any) response.FieldError {
	name := err.Field
	if i := strings.LastIndex(name, "."); i != -1 {
		name = name[i+1:]
	}
	label := fieldLabel(fieldByJSONName(obj, err.Field), name)

	t := err.Type
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	var message string
	switch t.Kind() {
	case reflect.String:
		message = fmt.Sprintf("%s must be text", label)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		message = fmt.Sprintf("%s must be a number", label)
	case reflect.Bool:
		message = fmt.Sprintf("%s must be true or false", label)
	case reflect.Slice, reflect.Array:
		message = fmt.Sprintf("%s must be a list", label)
	default:
		message = fmt.Sprintf("%s is invalid", label)
	}
	return response.FieldError{Code: "INVALID_TYPE", Field: err.Field, Message: message}
}

func messageFor(fe validator.FieldError, field *reflect.StructField, label string) string {
	if field != nil {
		// Per-rule override, for when the default wording doesn't fit:
		// msg_required:"Reset link is missing". Most fields need none.
		if override := strings.TrimSpace(field.Tag.Get("msg_" + fe.Tag())); override != "" {
			return override
		}
	}
	return buildMessage(fe, label)
}

func mapTagToCode(tag string) string {
	switch tag {
	case "required":
		return "REQUIRED_FIELD"
	case "min":
		return "TOO_SHORT"
	case "max", "max_bytes":
		return "TOO_LONG"
	case "email":
		return "INVALID_EMAIL"
	case "uuid":
		return "INVALID_UUID"
	case "url":
		return "INVALID_URL"
	case "person_name":
		return "INVALID_FORMAT"
	case "alphanum":
		return "INVALID_FORMAT"
	case "gt", "gte", "lt", "lte":
		return "OUT_OF_RANGE"
	case "oneof":
		return "INVALID_VALUE"
	case "boolean":
		return "INVALID_TYPE"
	case "numeric":
		return "INVALID_TYPE"
	case "dive":
		return "INVALID_ITEM"
	case "required_if":
		return "REQUIRED_FIELD"
	case "not_blank":
		return "BLANK_FIELD"
	case "date_format":
		return "INVALID_DATE"
	default:
		return "VALIDATION_ERROR"
	}
}

func buildMessage(fe validator.FieldError, label string) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", label)
	case "min":
		if fe.Kind() == reflect.Slice {
			return fmt.Sprintf("%s needs at least %s", label, items(fe.Param()))
		}
		return fmt.Sprintf("%s must be at least %s characters", label, fe.Param())
	case "max":
		if fe.Kind() == reflect.Slice {
			return fmt.Sprintf("%s can have at most %s", label, items(fe.Param()))
		}
		return fmt.Sprintf("%s cannot exceed %s characters", label, fe.Param())
	case "email":
		return "Enter a valid email address"
	case "uuid":
		return fmt.Sprintf("%s must be a valid ID", label)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", label)
	case "max_bytes":
		return fmt.Sprintf("%s is too long", label)
	case "person_name":
		return fmt.Sprintf("%s can only contain letters, spaces, and ' . -", label)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", label, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be at least %s", label, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", label, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be at most %s", label, fe.Param())
	case "oneof":
		// The allowed values are code (member, admin), not words to show a user.
		return fmt.Sprintf("Choose a valid %s", strings.ToLower(label))
	case "boolean":
		return fmt.Sprintf("%s must be true or false", label)
	case "numeric":
		return fmt.Sprintf("%s must be a number", label)
	case "alphanum":
		return fmt.Sprintf("%s can only contain letters and numbers", label)
	case "dive":
		return fmt.Sprintf("%s contains an invalid item", label)
	case "required_if":
		return fmt.Sprintf("%s is required", label)
	case "not_blank":
		return fmt.Sprintf("%s cannot be blank", label)
	case "date_format":
		return fmt.Sprintf("%s must be a valid date", label)
	default:
		return fmt.Sprintf("%s is invalid", label)
	}
}

// items formats a count of list items: "1 item", "3 items".
func items(n string) string {
	if n == "1" {
		return "1 item"
	}
	return n + " items"
}

func getStructField(obj any, fe validator.FieldError) *reflect.StructField {
	t := reflect.TypeOf(obj)
	if t == nil {
		return nil
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	field, ok := t.FieldByName(fe.StructField())
	if !ok {
		return nil
	}
	return &field
}

func jsonName(field *reflect.StructField, fe validator.FieldError) string {
	if field != nil {
		if tag := field.Tag.Get("json"); tag != "" {
			name := strings.Split(tag, ",")[0]
			if name != "" && name != "-" {
				return stripIndex(name)
			}
		}
	}

	name := fe.Field()
	if name == "" {
		name = fe.StructField()
	}
	return stripIndex(strings.ToLower(name))
}

// fieldByJSONName finds the top-level struct field whose json tag is name.
func fieldByJSONName(obj any, name string) *reflect.StructField {
	t := reflect.TypeOf(obj)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return nil
	}
	for i := range t.NumField() {
		field := t.Field(i)
		if strings.Split(field.Tag.Get("json"), ",")[0] == name {
			return &field
		}
	}
	return nil
}

func fieldLabel(field *reflect.StructField, jsonField string) string {
	if field != nil {
		if label := strings.TrimSpace(field.Tag.Get("label")); label != "" {
			return label
		}
	}
	return humanize(jsonField)
}

func humanize(name string) string {
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}

	runes := []rune(name)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func stripIndex(name string) string {
	if idx := strings.Index(name, "["); idx != -1 {
		return name[:idx]
	}
	return name
}
