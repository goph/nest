package nest

import (
	"encoding"
	"errors"
	"reflect"
)

// Decoder is implemented by types that can deserialize themselves.
type Decoder interface {
	Decode(value string) error
}

// canDecode checks whether a value can decode itself.
func canDecode(field reflect.Value) bool {
	// struct fields cannot fail this check
	if !field.CanInterface() {
		return false
	}

	_, ok := field.Interface().(Decoder)
	if !ok && field.CanAddr() {
		_, ok = field.Addr().Interface().(Decoder)
	}

	if !ok {
		_, ok = field.Interface().(encoding.TextUnmarshaler)
		if !ok && field.CanAddr() {
			_, ok = field.Addr().Interface().(encoding.TextUnmarshaler)
		}
	}

	return ok
}

// decode makes a value decode itself.
func decode(field reflect.Value, value string) error {
	if !canDecode(field) {
		return errors.New("value cannot decode itself")
	}

	d, ok := field.Interface().(Decoder)
	if !ok && field.CanAddr() {
		d, ok = field.Addr().Interface().(Decoder)
	}

	if ok {
		return d.Decode(value)
	}

	t, ok := field.Interface().(encoding.TextUnmarshaler)
	if !ok && field.CanAddr() {
		t, ok = field.Addr().Interface().(encoding.TextUnmarshaler)
	}

	if ok {
		return t.UnmarshalText([]byte(value))
	}

	return errors.New("failed to find a decoding type")
}
