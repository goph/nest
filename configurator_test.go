package nest_test

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/goph/nest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Decodable string

func (d *Decodable) Decode(value string) error {
	*d = Decodable(value)

	return nil
}

type DecodableStruct struct {
	Value string
}

func (d *DecodableStruct) Decode(value string) error {
	d.Value = value

	return nil
}

type Unmarshalable string

func (u *Unmarshalable) UnmarshalText(text []byte) error {
	*u = Unmarshalable(text)

	return nil
}

type UnmarshalableStruct struct {
	Value string
}

func (u *UnmarshalableStruct) UnmarshalText(text []byte) error {
	u.Value = string(text)

	return nil
}

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
	configurator.SetArgs([]string{"program", "--value", "value"})

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
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
	configurator.SetArgs([]string{"program", "--value", "value"})

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
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
	configurator.SetArgs([]string{"program", "--Value", "value"})

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConfigurator_Load_FlagHelp(t *testing.T) {
	type config struct {
		Value string `flag:""`
	}

	c := config{}

	var buf bytes.Buffer

	configurator := nest.NewConfigurator()
	configurator.SetArgs([]string{"program", "--help"})
	configurator.SetOutput(&buf)

	err := configurator.Load(&c)

	require.Error(t, err)
	assert.Equal(t, nest.ErrFlagHelp, err)
	assert.Equal(t, "Usage of program:\n\n\nFLAGS:\n\n      --value string   \n", buf.String())
}

func TestConfigurator_Load_FlagHelpWithName(t *testing.T) {
	type config struct {
		Value string `flag:""`
	}

	c := config{}

	var buf bytes.Buffer

	configurator := nest.NewConfigurator()
	configurator.SetName("my service")
	configurator.SetArgs([]string{"program", "--help"})
	configurator.SetOutput(&buf)

	err := configurator.Load(&c)

	require.Error(t, err)
	assert.Equal(t, nest.ErrFlagHelp, err)
	assert.Equal(t, "Usage of my service:\n\n\nFLAGS:\n\n      --value string   \n", buf.String())
}

func TestConfigurator_Load_FlagSplitWords(t *testing.T) {
	type SubConfig struct {
		Value string `flag:""`
	}

	type config struct {
		SubConfig `split_words:"true"`

		OtherValue string `flag:"" split_words:"true"`
		OtherSubConfig SubConfig `split_words:"true"`
	}

	expected := config{
		SubConfig: SubConfig{
			Value: "value",
		},

		OtherValue: "value",
		OtherSubConfig: SubConfig{
			Value: "value",
		},
	}
	actual := config{}

	configurator := nest.NewConfigurator()
	configurator.SetArgs([]string{"program", "--sub-config-value", "value", "--other-value", "value", "--other-sub-config-value", "value"})

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConfigurator_Load_FlagOsArgs(t *testing.T) {
	type config struct {
		Value string `flag:""`
	}

	expected := config{
		Value: "value",
	}
	actual := config{}

	backupArgs := os.Args
	os.Args = []string{"program", "--value", "value"}

	configurator := nest.NewConfigurator()

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Args = backupArgs
}

func TestConfigurator_Load_FlagEmpty(t *testing.T) {
	type config struct {
		Int   int   `flag:""`
		Int8  int8  `flag:""`
		Int32 int32 `flag:""`
		Int64 int64 `flag:""`

		Uint   uint   `flag:""`
		Uint8  uint8  `flag:""`
		Uint32 uint32 `flag:""`
		Uint64 uint64 `flag:""`

		Float32 float32 `flag:""`
		Float64 float64 `flag:""`

		Bool bool `flag:""`
	}

	expected := config{}
	actual := expected

	configurator := nest.NewConfigurator()
	configurator.SetArgs([]string{"program"})

	err := configurator.Load(&actual)

	require.NoError(t, err)
	assert.Equal(t, expected, actual)
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

func TestConfigurator_Load_EnvironmentSplitWords(t *testing.T) {
	type SubConfig struct {
		Value string `env:""`
	}

	type config struct {
		SubConfig `split_words:"true"`

		OtherValue string `env:"" split_words:"true"`
		OtherSubConfig SubConfig `split_words:"true"`
	}

	expected := config{
		SubConfig: SubConfig{
			Value: "value",
		},

		OtherValue: "value",
		OtherSubConfig: SubConfig{
			Value: "value",
		},
	}
	actual := config{}

	configurator := nest.NewConfigurator()

	os.Clearenv()
	os.Setenv("SUB_CONFIG_VALUE", "value")
	os.Setenv("OTHER_VALUE", "value")
	os.Setenv("OTHER_SUB_CONFIG_VALUE", "value")

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Clearenv()
}

func TestConfigurator_Load_EnvironmentEmpty(t *testing.T) {
	type config struct {
		Int   int   `env:""`
		Int8  int8  `env:""`
		Int32 int32 `env:""`
		Int64 int64 `env:""`

		Uint   uint   `env:""`
		Uint8  uint8  `env:""`
		Uint32 uint32 `env:""`
		Uint64 uint64 `env:""`

		Float32 float32 `env:""`
		Float64 float64 `env:""`

		Bool bool `env:""`
	}

	expected := config{}
	actual := expected

	configurator := nest.NewConfigurator()
	configurator.SetArgs([]string{"program"})

	err := configurator.Load(&actual)

	require.NoError(t, err)
	assert.Equal(t, expected, actual)
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

func TestConfigurator_Load_Struct(t *testing.T) {
	type subconfig struct {
		Value string `default:"default"`
	}

	type config struct {
		Sconfig subconfig
	}

	expected := config{
		Sconfig: subconfig{
			Value: "default",
		},
	}
	actual := config{}

	configurator := nest.NewConfigurator()

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConfigurator_Load_StructEnvWithPrefix(t *testing.T) {
	type subconfig struct {
		Value string `env:""`
	}

	type config struct {
		Sconfig subconfig
	}

	expected := config{
		Sconfig: subconfig{
			Value: "value",
		},
	}
	actual := config{}

	configurator := nest.NewConfigurator()
	configurator.SetEnvPrefix("app")

	os.Clearenv()
	os.Setenv("APP_SCONFIG_VALUE", "value")

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Clearenv()
}

func TestConfigurator_Load_Decodable(t *testing.T) {
	type subconfig struct {
		Value UnmarshalableStruct `default:"default"`
	}

	type config struct {
		Sconfig subconfig
	}

	expected := config{
		Sconfig: subconfig{
			Value: UnmarshalableStruct{
				Value: "default",
			},
		},
	}
	actual := config{}

	configurator := nest.NewConfigurator()

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConfigurator_Load_StructPrefixEnvWithPrefix(t *testing.T) {
	type subconfig struct {
		Value string `env:""`
	}

	type config struct {
		Sconfig subconfig `prefix:"subconfig"`
	}

	expected := config{
		Sconfig: subconfig{
			Value: "value",
		},
	}
	actual := config{}

	configurator := nest.NewConfigurator()
	configurator.SetEnvPrefix("app")

	os.Clearenv()
	os.Setenv("APP_SUBCONFIG_VALUE", "value")

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Clearenv()
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

		Duration time.Duration

		Decodable     Decodable
		Unmarshalable Unmarshalable
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

		Duration: 10 * time.Second,

		Decodable:     "value",
		Unmarshalable: "value",
	}
	actual := expected

	configurator := nest.NewConfigurator()

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConfigurator_Load_PointerTypes(t *testing.T) {
	type config struct {
		String *string `default:"string"`

		Int   *int   `default:"1"`
		Int8  *int8  `default:"1"`
		Int32 *int32 `default:"1"`
		Int64 *int64 `default:"1"`

		Uint   *uint   `default:"1"`
		Uint8  *uint8  `default:"1"`
		Uint32 *uint32 `default:"1"`
		Uint64 *uint64 `default:"1"`

		Float32 *float32 `default:"1.0"`
		Float64 *float64 `default:"1.0"`

		Bool *bool `default:"true"`

		Duration *time.Duration `default:"10s"`
	}

	var string = "string"

	var int int = 1
	var int8 int8 = 1
	var int32 int32 = 1
	var int64 int64 = 1

	var uint uint = 1
	var uint8 uint8 = 1
	var uint32 uint32 = 1
	var uint64 uint64 = 1

	var float32 float32 = 1.0
	var float64 float64 = 1.0

	var bool bool = true

	var duration time.Duration = 10 * time.Second

	expected := config{
		String: &string,

		Int:   &int,
		Int8:  &int8,
		Int32: &int32,
		Int64: &int64,

		Uint:   &uint,
		Uint8:  &uint8,
		Uint32: &uint32,
		Uint64: &uint64,

		Float32: &float32,
		Float64: &float64,

		Bool: &bool,

		Duration: &duration,
	}
	actual := config{}

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

		Duration time.Duration `default:"10s"`

		Decodable           Decodable           `default:"value"`
		DecodableStruct     DecodableStruct     `default:"value"`
		Unmarshalable       Unmarshalable       `default:"value"`
		UnmarshalableStruct UnmarshalableStruct `default:"value"`
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

		Duration: 10 * time.Second,

		Decodable: "value",
		DecodableStruct: DecodableStruct{
			Value: "value",
		},
		Unmarshalable: "value",
		UnmarshalableStruct: UnmarshalableStruct{
			Value: "value",
		},
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
		Flag:     "flag",
		Env:      "env",
		Default:  "default",
	}
	actual := config{
		Override: "override",
	}

	configurator := nest.NewConfigurator()
	configurator.SetArgs([]string{"program", "--flag", "flag"})

	os.Clearenv()
	os.Setenv("ENV", "env")

	err := configurator.Load(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Clearenv()
}

func TestConfigurator_Load_Help(t *testing.T) {
	type config struct {
		FlagValue string `flag:"value" default:"value" usage:"My flag value"`
		EnvValue  string `env:"value" default:"value" usage:"My env value"`
	}

	c := config{}

	var buf bytes.Buffer

	configurator := nest.NewConfigurator()
	configurator.SetArgs([]string{"program", "--help"})
	configurator.SetOutput(&buf)

	err := configurator.Load(&c)

	require.Error(t, err)
	assert.Equal(t, nest.ErrFlagHelp, err)
	assert.Equal(t, "Usage of program:\n\n\nFLAGS:\n\n      --value string   My flag value (default \"value\")\n\n\nENVIRONMENT VARIABLES:\n\n      VALUE string   My env value (default \"value\")\n", buf.String())
}
