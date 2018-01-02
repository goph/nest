package nest

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

type decodable struct {
	value string
}

func (d *decodable) Decode(value string) error {
	d.value = value

	return nil
}

func (d *decodable) getValue() string {
	return d.value
}

func TestCanDecode(t *testing.T) {
	tests := map[string]struct {
		v         interface{}
		decodable bool
	}{
		"decodable": {
			&decodable{},
			true,
		},
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
		"decodable": &decodable{},
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
