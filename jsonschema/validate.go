package jsonschema

import (
	"encoding/json"
	"errors"
)

func VerifySchemaAndUnmarshal(schema Definition, content []byte, v any) error {
	var data any
	err := json.Unmarshal(content, &data)
	if err != nil {
		return err
	}
	if !Validate(schema, data) {
		return errors.New("data validation failed against the provided schema")
	}
	return json.Unmarshal(content, &v)
}

func Validate(schema Definition, data any) bool {
	for _, dataType := range schema.Type {
		switch dataType {
		case Object:
			return validateObject(schema, data)
		case Array:
			return validateArray(schema, data)
		case String:
			_, ok := data.(string)
			return ok
		case Number: // float64 and int
			_, ok := data.(float64)
			if !ok {
				_, ok = data.(int)
			}
			return ok
		case Boolean:
			_, ok := data.(bool)
			return ok
		case Integer:
			// Golang unmarshals all numbers as float64, so we need to check if the float64 is an integer
			if num, ok := data.(float64); ok {
				return num == float64(int64(num))
			}
			_, ok := data.(int)
			return ok
		case Null:
			return data == nil
		default:
			return false
		}
	}

	return false
}

func validateObject(schema Definition, data any) bool {
	dataMap, ok := data.(map[string]any)
	if !ok {
		return false
	}
	for _, field := range schema.Required {
		if _, exists := dataMap[field]; !exists {
			return false
		}
	}
	for key, valueSchema := range schema.Properties {
		value, exists := dataMap[key]
		if exists && !Validate(valueSchema, value) {
			return false
		} else if !exists && contains(schema.Required, key) {
			return false
		}
	}
	return true
}

func validateArray(schema Definition, data any) bool {
	dataArray, ok := data.([]any)
	if !ok {
		return false
	}
	for _, item := range dataArray {
		if !Validate(*schema.Items, item) {
			return false
		}
	}
	return true
}

func contains[S ~[]E, E comparable](s S, v E) bool {
	for i := range s {
		if v == s[i] {
			return true
		}
	}
	return false
}
