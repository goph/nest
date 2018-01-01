package nest

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsZeroValueOfType(t *testing.T) {
	tests := map[string]interface{}{
		"string":   string(""),
		"rune":     rune('\x00'),
		"int":      int(0),
		"int32":    int32(0),
		"int64":    int64(0),
		"float32":  float32(0),
		"float64":  float64(0),
		"duration": time.Duration(0),
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if isZeroValueOfType(test) == false {
				t.Error("zero value not detected")
			}
		})
	}
}

func TestIsTrue(t *testing.T) {
	tests := map[string]bool{
		"true":            true,
		"TRUE":            true,
		"false":           false,
		"FALSE":           false,
		"oiajdfoidahfios": false,
	}

	for input, expected := range tests {
		t.Run("", func(t *testing.T) {
			if isTrue(input) != expected {
				t.Errorf("%s is expected to be parsed into %v, received %v", input, expected, !expected)
			}
		})
	}
}

func TestLowerFirst(t *testing.T) {
	tests := map[string]string{
		"string": "string",
		"STRING": "sTRING",
		"sTRING": "sTRING",
		"String": "string",
	}

	for input, expected := range tests {
		t.Run("", func(t *testing.T) {
			if actual := lowerFirst(input); actual != expected {
				t.Errorf("%s is expected to be parsed into %s, received %s", input, expected, actual)
			}
		})
	}
}

func TestSplitWords_Snake(t *testing.T) {
	tests := map[string]string{
		"CamelCase": "camel_case",
		"camelCase": "camel_case",
		"camel":     "camel",
	}

	for input, expected := range tests {
		t.Run("", func(t *testing.T) {
			if actual := splitWords(input, "_"); actual != expected {
				t.Errorf("%s is expected to become %s, received %s", input, expected, actual)
			}
		})
	}
}

func TestSplitWords_Spinal(t *testing.T) {
	tests := map[string]string{
		"CamelCase": "camel-case",
		"camelCase": "camel-case",
		"camel":     "camel",
	}

	for input, expected := range tests {
		t.Run("", func(t *testing.T) {
			if actual := splitWords(input, "-"); actual != expected {
				t.Errorf("%s is expected to become %s, received %s", input, expected, actual)
			}
		})
	}
}

func TestIsExported(t *testing.T) {
	tests := map[string]bool{
		"nonExported": false,
		"Exported":    true,
	}

	for name, result := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, result, isExported(name))
		})
	}
}

type unmarshalable struct {
	value string
}

func (u *unmarshalable) UnmarshalText(text []byte) error {
	u.value = string(text)

	return nil
}

func (u *unmarshalable) getValue() string {
	return u.value
}

func TestCanDecode(t *testing.T) {
	tests := map[string]struct {
		v         interface{}
		decodable bool
	}{
		"unmarshalable": {
			&unmarshalable{},
			true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.decodable, canDecode(reflect.ValueOf(test.v)))
		})
	}
}

func TestDecode(t *testing.T) {
	tests := map[string]interface {
		getValue() string
	}{
		"unmarshalable": &unmarshalable{},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			field := reflect.ValueOf(test)

			if assert.True(t, canDecode(field)) {
				err := decode(field, "data")
				require.NoError(t, err)
				assert.Equal(t, "data", test.getValue())
			}
		})
	}
}
