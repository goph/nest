package nest_test

import (
	"testing"
	"time"

	"github.com/goph/nest"
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
			if nest.IsZeroValueOfType(test) == false {
				t.Error("zero value not detected")
			}
		})
	}
}
