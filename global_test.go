package nest_test

import (
	"os"
	"testing"

	"github.com/goph/nest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_EnvironmentWithPrefix(t *testing.T) {
	type config struct {
		Value string `env:""`
	}

	expected := config{
		Value: "value",
	}
	actual := config{}

	nest.SetEnvPrefix("app")

	os.Clearenv()
	os.Setenv("APP_VALUE", "value")

	err := nest.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Clearenv()
}
