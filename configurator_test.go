package nest_test

import (
	"os"
	"testing"

	"github.com/goph/nest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurator_Load_NotStructPointer(t *testing.T) {
	type config struct {
		Value string
	}

	c := config{}

	configurator := nest.NewConfigurator()

	err := configurator.Load(c)
	require.Error(t, err)
	assert.Equal(t, nest.ErrNotStructPointer, err)
}

func TestConfigurator_Load_NotStruct(t *testing.T) {
	var c string

	configurator := nest.NewConfigurator()

	err := configurator.Load(&c)
	require.Error(t, err)
	assert.Equal(t, nest.ErrNotStruct, err)
}

func TestConfigurator_Load_IgnoreUnexportedField(t *testing.T) {
	type config struct {
		value string `default:"default"`
	}

	expected := config{}
	actual := expected

	configurator := nest.NewConfigurator()

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConfigurator_Load_Ignored(t *testing.T) {
	type config struct {
		Value string `ignored:"true" default:"default"`
	}

	expected := config{}
	actual := expected

	configurator := nest.NewConfigurator()

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConfigurator_Load_Required(t *testing.T) {
	type config struct {
		Value string `required:"true"`
	}

	c := config{}

	configurator := nest.NewConfigurator()

	err := configurator.Load(&c)
	require.Error(t, err)
	assert.EqualError(t, err, "required field Value missing value")
}

func TestConfigurator_Load_RequiredWithDefault(t *testing.T) {
	type config struct {
		Value string `required:"true" default:"default"`
	}

	expected := config{
		Value: "default",
	}
	actual := config{}

	configurator := nest.NewConfigurator()

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConfigurator_Load_Overrides(t *testing.T) {
	type config struct {
		Value string
	}

	expected := config{
		Value: "value",
	}
	actual := expected

	configurator := nest.NewConfigurator()

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConfigurator_Load_Flag(t *testing.T) {
	type config struct {
		Value string `flag:""`
	}

	expected := config{
		Value: "value",
	}
	actual := config{}

	configurator := nest.NewConfigurator()

	backupArgs := os.Args
	os.Args = []string{"program", "--value", "value"}

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Args = backupArgs
}

func TestConfigurator_Load_FlagWithAlias(t *testing.T) {
	type config struct {
		Value string `flag:"value"`
	}

	expected := config{
		Value: "value",
	}
	actual := config{}

	configurator := nest.NewConfigurator()

	backupArgs := os.Args
	os.Args = []string{"program", "--value", "value"}

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Args = backupArgs
}

func TestConfigurator_Load_FlagWithUpperCaseFirstAlias(t *testing.T) {
	type config struct {
		Value string `flag:"Value"`
	}

	expected := config{
		Value: "value",
	}
	actual := config{}

	configurator := nest.NewConfigurator()

	backupArgs := os.Args
	os.Args = []string{"program", "--Value", "value"}

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Args = backupArgs
}

func TestConfigurator_Load_Environment(t *testing.T) {
	type config struct {
		Value string `env:""`
	}

	expected := config{
		Value: "value",
	}
	actual := config{}

	configurator := nest.NewConfigurator()

	os.Clearenv()
	os.Setenv("VALUE", "value")

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Clearenv()
}

func TestConfigurator_Load_EnvironmentWithAlias(t *testing.T) {
	type config struct {
		Value string `env:"other_value"`
	}

	expected := config{
		Value: "value",
	}
	actual := config{}

	configurator := nest.NewConfigurator()

	os.Clearenv()
	os.Setenv("OTHER_VALUE", "value")

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Clearenv()
}

func TestConfigurator_Load_EnvironmentWithPrefix(t *testing.T) {
	type config struct {
		Value string `env:""`
	}

	expected := config{
		Value: "value",
	}
	actual := config{}

	configurator := nest.NewConfigurator()
	configurator.SetEnvPrefix("app")

	os.Clearenv()
	os.Setenv("APP_VALUE", "value")

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Clearenv()
}

func TestConfigurator_Load_EnvironmentWithPrefixAndAlias(t *testing.T) {
	type config struct {
		Value string `env:"other_value"`
	}

	expected := config{
		Value: "value",
	}
	actual := config{}

	configurator := nest.NewConfigurator()
	configurator.SetEnvPrefix("app")

	os.Clearenv()
	os.Setenv("APP_OTHER_VALUE", "value")

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Clearenv()
}

func TestConfigurator_Load_Default(t *testing.T) {
	type config struct {
		Value string `default:"default"`
	}

	expected := config{
		Value: "default",
	}
	actual := config{}

	configurator := nest.NewConfigurator()

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConfigurator_Load_Types(t *testing.T) {
	type config struct {
		String string

		Int   int
		Int8  int8
		Int32 int32
		Int64 int64

		Uint   uint
		Uint8  uint8
		Uint32 uint32
		Uint64 uint64

		Float32 float32
		Float64 float64

		Bool bool
	}

	expected := config{
		String: "string",

		Int:   1,
		Int8:  1,
		Int32: 1,
		Int64: 1,

		Uint:   1,
		Uint8:  1,
		Uint32: 1,
		Uint64: 1,

		Float32: 1.0,
		Float64: 1.0,

		Bool: true,
	}
	actual := expected

	configurator := nest.NewConfigurator()

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConfigurator_Load_TypeDefaults(t *testing.T) {
	type config struct {
		String string `default:"string"`

		Int   int   `default:"1"`
		Int8  int8  `default:"1"`
		Int32 int32 `default:"1"`
		Int64 int64 `default:"1"`

		Uint   uint   `default:"1"`
		Uint8  uint8  `default:"1"`
		Uint32 uint32 `default:"1"`
		Uint64 uint64 `default:"1"`

		Float32 float32 `default:"1.0"`
		Float64 float64 `default:"1.0"`

		Bool bool `default:"true"`
	}

	expected := config{
		String: "string",

		Int:   1,
		Int8:  1,
		Int32: 1,
		Int64: 1,

		Uint:   1,
		Uint8:  1,
		Uint32: 1,
		Uint64: 1,

		Float32: 1.0,
		Float64: 1.0,

		Bool: true,
	}
	actual := config{}

	configurator := nest.NewConfigurator()

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConfigurator_Load_PrecedenceOrder(t *testing.T) {
	type config struct {
		Override string `env:"" flag:"" default:"default"`
		Flag     string `env:"" flag:"" default:"default"`
		Env      string `env:"" flag:"" default:"default"`
		Default  string `default:"default"`
	}

	expected := config{
		Override: "override",
		Flag:     "value",
		Env:      "value",
		Default:  "default",
	}
	actual := config{
		Override: "override",
	}

	configurator := nest.NewConfigurator()

	os.Clearenv()
	os.Setenv("ENV", "value")
	backupArgs := os.Args
	os.Args = []string{"program", "--flag", "value"}

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Clearenv()
	os.Args = backupArgs
}
