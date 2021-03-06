package nest

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestField_IgnoreUnexportedField(t *testing.T) {
	type config struct {
		value string `required:"true"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	var expected []fieldDefinition

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_Ignored(t *testing.T) {
	type config struct {
		Value string `ignored:"true"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	var expected []fieldDefinition

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_Required(t *testing.T) {
	type config struct {
		Value string `required:"true"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Value",
			field: ref.Field(0),

			required: true,
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_Overrides(t *testing.T) {
	type config struct {
		Value string
	}

	c := config{
		Value: "value",
	}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Value",
			field: ref.Field(0),

			hasOverride:   true,
			overrideValue: "value",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_Flag(t *testing.T) {
	type config struct {
		Value string `flag:""`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Value",
			field: ref.Field(0),

			hasFlag:   true,
			flagAlias: "value",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_FlagWithAlias(t *testing.T) {
	type config struct {
		Value string `flag:"value"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Value",
			field: ref.Field(0),

			hasFlag:   true,
			flagAlias: "value",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_FlagWithUpperCaseFirstAlias(t *testing.T) {
	type config struct {
		Value string `flag:"Value"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Value",
			field: ref.Field(0),

			hasFlag:   true,
			flagAlias: "Value",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_FlagSplitWords(t *testing.T) {
	type config struct {
		OtherValue string `flag:"" split_words:"true"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "OtherValue",
			field: ref.Field(0),

			hasFlag:   true,
			flagAlias: "other-value",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_Environment(t *testing.T) {
	type config struct {
		Value string `env:""`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Value",
			field: ref.Field(0),

			hasEnv:   true,
			envAlias: "VALUE",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_EnvironmentWithAlias(t *testing.T) {
	type config struct {
		Value string `env:"other_value"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Value",
			field: ref.Field(0),

			hasEnv:   true,
			envAlias: "OTHER_VALUE",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_EnvironmentSplitWords(t *testing.T) {
	type config struct {
		OtherValue string `env:"" split_words:"true"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "OtherValue",
			field: ref.Field(0),

			hasEnv:   true,
			envAlias: "OTHER_VALUE",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_Default(t *testing.T) {
	type config struct {
		Value string `default:"default"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Value",
			field: ref.Field(0),

			hasDefault:   true,
			defaultValue: "default",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_ChildStruct(t *testing.T) {
	type subconfig struct {
		Value string `default:"default"`
	}

	type config struct {
		Sconfig subconfig
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Sconfig.Value",
			field: ref.Field(0).Field(0),

			hasDefault:   true,
			defaultValue: "default",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_ChildStruct_Prefix(t *testing.T) {
	type subconfig struct {
		Value string `default:"default"`
	}

	type config struct {
		Sconfig subconfig `prefix:"subconfig"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "subconfig.Value",
			field: ref.Field(0).Field(0),

			hasDefault:   true,
			defaultValue: "default",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_ChildStruct_Flag(t *testing.T) {
	type subconfig struct {
		Value string `flag:""`
	}

	type config struct {
		Sconfig subconfig
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Sconfig.Value",
			field: ref.Field(0).Field(0),

			hasFlag:   true,
			flagAlias: "sconfig-value",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_ChildStruct_Prefix_Flag(t *testing.T) {
	type subconfig struct {
		Value string `flag:""`
	}

	type config struct {
		Sconfig subconfig `prefix:"subconfig"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "subconfig.Value",
			field: ref.Field(0).Field(0),

			hasFlag:   true,
			flagAlias: "subconfig-value",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_ChildStructMulti_Flag(t *testing.T) {
	type subsubconfig struct {
		Value string `flag:""`
	}

	type subconfig struct {
		Sconfig subsubconfig
	}

	type config struct {
		Sconfig subconfig
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Sconfig.Sconfig.Value",
			field: ref.Field(0).Field(0).Field(0),

			hasFlag:   true,
			flagAlias: "sconfig-sconfig-value",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_ChildStructMulti_Prefix_Flag(t *testing.T) {
	type subsubconfig struct {
		Value string `flag:""`
	}

	type subconfig struct {
		Sconfig subsubconfig `prefix:"subconfig2"`
	}

	type config struct {
		Sconfig subconfig `prefix:"subconfig"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "subconfig.subconfig2.Value",
			field: ref.Field(0).Field(0).Field(0),

			hasFlag:   true,
			flagAlias: "subconfig-subconfig2-value",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_ChildStruct_Environment(t *testing.T) {
	type subconfig struct {
		Value string `env:""`
	}

	type config struct {
		Sconfig subconfig
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Sconfig.Value",
			field: ref.Field(0).Field(0),

			hasEnv:   true,
			envAlias: "SCONFIG_VALUE",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_ChildStruct_Prefix_Environment(t *testing.T) {
	type subconfig struct {
		Value string `env:""`
	}

	type config struct {
		Sconfig subconfig `prefix:"subconfig"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "subconfig.Value",
			field: ref.Field(0).Field(0),

			hasEnv:   true,
			envAlias: "SUBCONFIG_VALUE",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_ChildStruct_EnvironmentWithAlias(t *testing.T) {
	type subconfig struct {
		Value string `env:"other_value"`
	}

	type config struct {
		Sconfig subconfig
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Sconfig.Value",
			field: ref.Field(0).Field(0),

			hasEnv:   true,
			envAlias: "SCONFIG_OTHER_VALUE",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_ChildStructMulti_EnvironmentWithAlias(t *testing.T) {
	type subsubconfig struct {
		Value string `env:"other_value"`
	}

	type subconfig struct {
		Sconfig subsubconfig
	}

	type config struct {
		Sconfig subconfig
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Sconfig.Sconfig.Value",
			field: ref.Field(0).Field(0).Field(0),

			hasEnv:   true,
			envAlias: "SCONFIG_SCONFIG_OTHER_VALUE",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_ChildStructMulti_Prefix_EnvironmentWithAlias(t *testing.T) {
	type subsubconfig struct {
		Value string `env:"other_value"`
	}

	type subconfig struct {
		Sconfig subsubconfig `prefix:"subconfig2"`
	}

	type config struct {
		Sconfig subconfig `prefix:"subconfig"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "subconfig.subconfig2.Value",
			field: ref.Field(0).Field(0).Field(0),

			hasEnv:   true,
			envAlias: "SUBCONFIG_SUBCONFIG2_OTHER_VALUE",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_EmbeddedStruct(t *testing.T) {
	type Subconfig struct {
		Value string `default:"default"`
	}

	type config struct {
		Subconfig
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Subconfig.Value",
			field: ref.Field(0).Field(0),

			hasDefault:   true,
			defaultValue: "default",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_EmbeddedStruct_Prefix(t *testing.T) {
	type Subconfig struct {
		Value string `default:"default"`
	}

	type config struct {
		Subconfig `prefix:"subconfig"`

		OtherSubConfig struct {
			Subconfig `prefix:""`
		} `prefix:"other"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "subconfig.Value",
			field: ref.Field(0).Field(0),

			hasDefault:   true,
			defaultValue: "default",
		},
		{
			key:   "other.Value",
			field: ref.Field(1).Field(0).Field(0),

			hasDefault:   true,
			defaultValue: "default",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_EmbeddedStruct_Flag(t *testing.T) {
	type Subconfig struct {
		Value string `flag:""`
	}

	type config struct {
		Subconfig
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Subconfig.Value",
			field: ref.Field(0).Field(0),

			hasFlag:   true,
			flagAlias: "subconfig-value",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_EmbeddedStruct_Prefix_Flag(t *testing.T) {
	type Subconfig struct {
		Value string `flag:""`
	}

	type config struct {
		Subconfig `prefix:"subconfig"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "subconfig.Value",
			field: ref.Field(0).Field(0),

			hasFlag:   true,
			flagAlias: "subconfig-value",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_EmbeddedStructMulti_Flag(t *testing.T) {
	type Subsubconfig struct {
		Value string `flag:""`
	}

	type Subconfig struct {
		Subsubconfig
	}

	type config struct {
		Subconfig
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Subconfig.Subsubconfig.Value",
			field: ref.Field(0).Field(0).Field(0),

			hasFlag:   true,
			flagAlias: "subconfig-subsubconfig-value",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_EmbeddedStructMulti_Prefix_Flag(t *testing.T) {
	type Subsubconfig struct {
		Value string `flag:""`
	}

	type Subconfig struct {
		Subsubconfig `prefix:"subconfig2"`
	}

	type config struct {
		Subconfig `prefix:"subconfig"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "subconfig.subconfig2.Value",
			field: ref.Field(0).Field(0).Field(0),

			hasFlag:   true,
			flagAlias: "subconfig-subconfig2-value",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_EmbeddedStruct_Environment(t *testing.T) {
	type Subconfig struct {
		Value string `env:""`
	}

	type config struct {
		Subconfig
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Subconfig.Value",
			field: ref.Field(0).Field(0),

			hasEnv:   true,
			envAlias: "SUBCONFIG_VALUE",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_EmbeddedStruct_Prefix_Environment(t *testing.T) {
	type Subconfig struct {
		Value string `env:""`
	}

	type config struct {
		Subconfig `prefix:"subconfig"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "subconfig.Value",
			field: ref.Field(0).Field(0),

			hasEnv:   true,
			envAlias: "SUBCONFIG_VALUE",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_EmbeddedStruct_EnvironmentWithAlias(t *testing.T) {
	type Subconfig struct {
		Value string `env:"other_value"`
	}

	type config struct {
		Subconfig
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Subconfig.Value",
			field: ref.Field(0).Field(0),

			hasEnv:   true,
			envAlias: "SUBCONFIG_OTHER_VALUE",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_EmbeddedStructMulti_EnvironmentWithAlias(t *testing.T) {
	type Subsubconfig struct {
		Value string `env:"other_value"`
	}

	type Subconfig struct {
		Subsubconfig
	}

	type config struct {
		Subconfig
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "Subconfig.Subsubconfig.Value",
			field: ref.Field(0).Field(0).Field(0),

			hasEnv:   true,
			envAlias: "SUBCONFIG_SUBSUBCONFIG_OTHER_VALUE",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_EmbeddedStruct_Prefix_EnvironmentWithAlias(t *testing.T) {
	type Subconfig struct {
		Value string `env:"other_value"`
	}

	type config struct {
		Subconfig `prefix:"subconfig"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "subconfig.Value",
			field: ref.Field(0).Field(0),

			hasEnv:   true,
			envAlias: "SUBCONFIG_OTHER_VALUE",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_EmbeddedStructMulti_Prefix_EnvironmentWithAlias(t *testing.T) {
	type Subsubconfig struct {
		Value string `env:"other_value"`
	}

	type Subconfig struct {
		Subsubconfig `prefix:"subconfig2"`
	}

	type config struct {
		Subconfig `prefix:"subconfig"`
	}

	c := config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "subconfig.subconfig2.Value",
			field: ref.Field(0).Field(0).Field(0),

			hasEnv:   true,
			envAlias: "SUBCONFIG_SUBCONFIG2_OTHER_VALUE",
		},
	}

	actual := getDefinitions(ref)
	assert.Equal(t, expected, actual)
}

func TestField_Decode(t *testing.T) {
	type config struct {
		DecodableValue     Decodable     `default:"default"`
		UnmarshalableValue Unmarshalable `default:"default"`
	}

	c := &config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "DecodableValue",
			field: ref.Elem().Field(0),

			hasDefault:   true,
			defaultValue: "default",
		},
		{
			key:   "UnmarshalableValue",
			field: ref.Elem().Field(1),

			hasDefault:   true,
			defaultValue: "default",
		},
	}

	actual := getDefinitions(ref.Elem())
	assert.Equal(t, expected, actual)
}

func TestField_StructDecode(t *testing.T) {
	type config struct {
		DecodableValue     DecodableStruct     `default:"default"`
		UnmarshalableValue UnmarshalableStruct `default:"default"`
	}

	c := &config{}
	ref := reflect.ValueOf(c)
	expected := []fieldDefinition{
		{
			key:   "DecodableValue",
			field: ref.Elem().Field(0),

			hasDefault:   true,
			defaultValue: "default",
		},
		{
			key:   "UnmarshalableValue",
			field: ref.Elem().Field(1),

			hasDefault:   true,
			defaultValue: "default",
		},
	}

	actual := getDefinitions(ref.Elem())
	assert.Equal(t, expected, actual)
}
