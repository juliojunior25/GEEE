package base

import (
	"fmt"
	"reflect"
)

// ValidateAgainstSchema validates data against a JSON schema
// This is a simplified implementation - for production use a full JSON Schema validator
func ValidateAgainstSchema(data map[string]interface{}, schema map[string]interface{}) error {
	// Get required fields
	required, _ := schema["required"].([]interface{})
	requiredFields := make(map[string]bool)
	for _, field := range required {
		if fieldName, ok := field.(string); ok {
			requiredFields[fieldName] = true
		}
	}

	// Check required fields
	for fieldName := range requiredFields {
		if _, exists := data[fieldName]; !exists {
			return fmt.Errorf("required field '%s' is missing", fieldName)
		}
	}

	// Get properties schema
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return nil // No properties to validate
	}

	// Validate each field
	for fieldName, fieldValue := range data {
		fieldSchema, hasSchema := properties[fieldName]
		if !hasSchema {
			// Field not in schema - check if additionalProperties is allowed
			if additionalProps, ok := schema["additionalProperties"].(bool); ok && !additionalProps {
				return fmt.Errorf("field '%s' is not allowed by schema", fieldName)
			}
			continue
		}

		// Validate field type
		if err := validateFieldType(fieldName, fieldValue, fieldSchema); err != nil {
			return err
		}
	}

	return nil
}

// validateFieldType validates a field's type
func validateFieldType(fieldName string, value interface{}, schema interface{}) error {
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return nil // No type validation needed
	}

	expectedType, ok := schemaMap["type"].(string)
	if !ok {
		return nil // No type specified
	}

	actualType := getJSONType(value)

	if expectedType != actualType {
		return fmt.Errorf("field '%s' has type '%s', expected '%s'", fieldName, actualType, expectedType)
	}

	return nil
}

// getJSONType returns the JSON type of a value
func getJSONType(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return "unknown"
	}
}
