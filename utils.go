package nest

import (
	"encoding"
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var splitWordsRegexp *regexp.Regexp

func init() {
	splitWordsRegexp = regexp.MustCompile("([^A-Z]+|[A-Z][^A-Z]+|[A-Z]+)")
}

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

// splitWords splits a camel cased string and converts it to snake or spinal case (according to the glue string).
func splitWords(s string, glue string) string {
	words := splitWordsRegexp.FindAllStringSubmatch(s, -1)
	if len(words) < 1 {
		return ""
	}

	var name []string
	for _, words := range words {
		name = append(name, words[0])
	}

	return strings.ToLower(strings.Join(name, glue))
}

// isExported checks whether a struct field is exported or not.
func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)

	return unicode.IsUpper(r)
}

// canDecode checks whether a value can decode itself.
func canDecode(field reflect.Value) bool {
	// struct fields cannot fail this check
	if !field.CanInterface() {
		return false
	}

	_, ok := field.Interface().(encoding.TextUnmarshaler)
	if !ok && field.CanAddr() {
		_, ok = field.Addr().Interface().(encoding.TextUnmarshaler)
	}

	return ok
}

// decode makes a value decode itself.
func decode(field reflect.Value, value string) error {
	if !canDecode(field) {
		return errors.New("value cannot decode itself")
	}

	v, ok := field.Interface().(encoding.TextUnmarshaler)
	if !ok && field.CanAddr() {
		v, ok = field.Addr().Interface().(encoding.TextUnmarshaler)
	}

	if !ok {
		return errors.New("failed to find a decoding type")
	}

	return v.UnmarshalText([]byte(value))
}
