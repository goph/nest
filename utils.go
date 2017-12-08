package nest

import "reflect"

// IsZeroValueOfType checks whether an interface typed value holds the zero value of the underlying type.
//
// Source: https://stackoverflow.com/a/13906031/3027614
func IsZeroValueOfType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
