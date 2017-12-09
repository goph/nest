package nest

import (
	"reflect"
	"strconv"
	"unicode"
)

// isZeroValueOfType checks whether an interface typed value holds the zero value of the underlying type.
//
// Source: https://stackoverflow.com/a/13906031/3027614
func isZeroValueOfType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

// isTrue checks whether a string contains a value which can be parsed into "true" boolean value.
func isTrue(s string) bool {
	b, _ := strconv.ParseBool(s)

	return b
}

// lowerFirst converts the first character of a string to lower case.
func lowerFirst(s string) string {
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])

	return string(a)
}
