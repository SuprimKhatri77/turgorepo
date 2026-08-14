package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
)

func Parse(err error, obj any) []types.AppError {
	var ve validator.ValidationErrors

	if !errors.As(err, &ve) {
		return []types.AppError{
			{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		}
	}

	var errs []types.AppError

	for _, fe := range ve {
		field := getStructField(obj, fe)
		jsonField := jsonName(field, fe)
		label := fieldLabel(field, jsonField)

		errs = append(errs, types.AppError{
			Code:    mapTagToCode(fe.Tag()),
			Field:   jsonField,
			Message: messageFor(fe, field, label),
		})
	}

	return errs
}

func messageFor(fe validator.FieldError, field *reflect.StructField, label string) string {
	if field != nil {
		// Per-rule override: msg_required, msg_min, msg_email, ...
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
	case "max":
		return "TOO_LONG"
	case "email":
		return "INVALID_EMAIL"
	case "uuid":
		return "INVALID_UUID"
	case "url":
		return "INVALID_URL"
	case "alphaspace":
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
			return fmt.Sprintf("%s must have at least %s item(s)", label, fe.Param())
		}
		return fmt.Sprintf("%s must be at least %s characters", label, fe.Param())
	case "max":
		if fe.Kind() == reflect.Slice {
			return fmt.Sprintf("%s cannot have more than %s item(s)", label, fe.Param())
		}
		return fmt.Sprintf("%s cannot exceed %s characters", label, fe.Param())
	case "email":
		return fmt.Sprintf("Enter a valid %s", strings.ToLower(label))
	case "uuid":
		return fmt.Sprintf("%s must be a valid ID", label)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", label)
	case "alphaspace":
		return fmt.Sprintf("%s can only contain letters and spaces", label)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", label, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be at least %s", label, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", label, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be at most %s", label, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", label, strings.ReplaceAll(fe.Param(), " ", ", "))
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
		return fmt.Sprintf("%s must be a valid date (YYYY-MM-DD)", label)
	default:
		return fmt.Sprintf("%s is invalid", label)
	}
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
