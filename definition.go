package nest

import (
	"go/ast"
	"reflect"
	"strings"
)

type fieldDefinition struct {
	key   string
	field reflect.Value

	hasOverride   bool
	overrideValue interface{}

	hasFlag   bool
	flagAlias string

	hasEnv   bool
	envAlias string

	hasDefault   bool
	defaultValue string

	required bool
}

func getDefinitions(structRef reflect.Value) []fieldDefinition {
	return getDefinitionsForStruct(structRef, "")
}

func getDefinitionsForStruct(structRef reflect.Value, prefix string) []fieldDefinition {
	structType := structRef.Type()

	var keyPrefix string
	if prefix != "" {
		keyPrefix = prefix + "."
	}

	var flagPrefix string
	if prefix != "" {
		flagPrefix = lowerFirst(prefix) + "-"
	}

	var envPrefix string
	if prefix != "" {
		envPrefix = strings.ToLower(prefix) + "_"
	}

	var definitions []fieldDefinition

	// Gather configuration definition information
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		field := structRef.Field(i)

		// Ignore unexported field
		if ast.IsExported(structField.Name) == false {
			continue
		}

		// Manually ignored field
		if value, ok := structField.Tag.Lookup(TagIgnored); ok && isTrue(value) {
			continue
		}

		// Resolve pointer to it's actual type
		for field.Kind() == reflect.Ptr {
			// Set to zero value when field is nil
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}

			field = field.Elem()
		}

		// Process child struct fields
		if field.Kind() == reflect.Struct {
			structDefinitions := getDefinitionsForStruct(field, structField.Name)
			definitions = append(definitions, structDefinitions...)

			continue
		}

		// Ignore unsupported field
		if _, unsupported := unsupportedTypes[field.Kind()]; unsupported {
			continue
		}

		def := fieldDefinition{
			key:   keyPrefix + structField.Name,
			field: field,
		}

		// Set value override
		if value := field.Interface(); isZeroValueOfType(value) == false {
			def.hasOverride = true
			def.overrideValue = value
		}

		// Map flag to field
		if value, ok := structField.Tag.Lookup(TagFlag); ok {
			def.hasFlag = true

			// Use the field name as flag name if it is not provided
			if value == "" {
				// Make the first character lower case, because that's customary
				value = lowerFirst(structField.Name)

				// Try to split words in the struct name if possible
				if v, ok := structField.Tag.Lookup(TagSplitWords); ok && isTrue(v) {
					v = splitWords(value, "-")
					if v != "" {
						value = v
					}
				}
			}

			def.flagAlias = flagPrefix + value
		}

		// Map environment variable to field
		if value, ok := structField.Tag.Lookup(TagEnvironment); ok {
			def.hasEnv = true

			// An environment variable alias is provided
			if value != "" {
				def.envAlias = strings.ToUpper(envPrefix + value)
			} else if v, ok := structField.Tag.Lookup(TagSplitWords); ok && isTrue(v) { // Try to split words in the struct name if possible
				v = splitWords(structField.Name, "_")
				if v != "" {
					def.envAlias = strings.ToUpper(envPrefix + v)
				}
			}
		}

		// Set default (if any)
		if value, ok := structField.Tag.Lookup(TagDefault); ok {
			def.hasDefault = true
			def.defaultValue = value
		}

		// Check if the field is required
		if value, ok := structField.Tag.Lookup(TagRequired); ok && isTrue(value) {
			def.required = true
		}

		definitions = append(definitions, def)
	}

	return definitions
}
