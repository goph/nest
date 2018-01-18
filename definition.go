package nest

import (
	"reflect"
	"strings"
)

// unsupportedTypes is a list of types that cannot be configured at the moment.
var unsupportedTypes = map[reflect.Kind]bool{
	reflect.Complex64:     true,
	reflect.Complex128:    true,
	reflect.Array:         true,
	reflect.Chan:          true,
	reflect.Func:          true,
	reflect.Interface:     true,
	reflect.Map:           true,
	reflect.Ptr:           true,
	reflect.Slice:         true,
	reflect.UnsafePointer: true,
}

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

	usage string
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
		flagPrefix = strings.ToLower(strings.Replace(prefix, ".", "-", -1)) + "-"
	}

	var envPrefix string
	if prefix != "" {
		envPrefix = strings.ToLower(strings.Replace(prefix, ".", "_", -1)) + "_"
	}

	var definitions []fieldDefinition

	// Gather configuration definition information
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		field := structRef.Field(i)

		// Ignore unexported field
		if isExported(structField.Name) == false {
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
		if field.Kind() == reflect.Struct && !canDecode(field) {
			value, ok := structField.Tag.Lookup(TagPrefix)
			prefix := keyPrefix + value

			// No prefix is provided, guess the prefix from the struct name
			if !ok {
				name := structField.Name

				// Try to split words in the struct name if possible
				if v, ok := structField.Tag.Lookup(TagSplitWords); ok && isTrue(v) {
					v = splitWords(name, ".")
					if v != "" {
						name = v
					}
				}

				prefix = keyPrefix + name
			}

			structDefinitions := getDefinitionsForStruct(field, prefix)
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

			usage: structField.Tag.Get(TagUsage),
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
			} else {
				def.envAlias = strings.ToUpper(envPrefix + structField.Name)
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
