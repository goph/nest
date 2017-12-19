package nest

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
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
