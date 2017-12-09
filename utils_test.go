package nest

import (
	"testing"
	"time"
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
