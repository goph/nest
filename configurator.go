package nest

import (
	"errors"
	"fmt"
	"go/ast"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// ErrNotStructPointer is returned when value passed to config.Load() is not a pointer to a struct.
	ErrNotStructPointer = errors.New("value passed is not a struct pointer")

	// ErrNotStruct is returned when value passed to config.Load() is not a struct.
	ErrNotStruct = errors.New("value passed is not a struct")

	ErrFlagHelp = pflag.ErrHelp
)

// unsupportedTypes is a list of types that cannot be configured at the moment.
var unsupportedTypes = map[reflect.Kind]bool{
	reflect.Complex64:     true,
	reflect.Complex128:    true,
	reflect.Array:         true,
	reflect.Chan:          true,
	reflect.Func:          true,
	reflect.Interface:     true,
	reflect.Map:           true,
	reflect.Ptr:           true,
	reflect.Slice:         true,
	reflect.Struct:        true,
	reflect.UnsafePointer: true,
}

func NewConfigurator() *Configurator {
	return &Configurator{
		args: os.Args,
		viper: viper.New(),
	}
}

type Configurator struct {
	// Used when displaying help
	name string

	// Command line arguments (defaults to os.Args)
	args []string

	// Environment prefix
	envPrefix string

	viper *viper.Viper

	mu sync.Mutex
}

// SetEnvPrefix manually sets the environment variable prefix.
func (c *Configurator) SetEnvPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.envPrefix = prefix
	c.viper.SetEnvPrefix(prefix)
}

// SetName sets the application name for displaying help.
func (c *Configurator) SetName(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.name = name
}

// SetArgs sets the command line arguments.
func (c *Configurator) SetArgs(args []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.args = args
}

// mergeWithEnvPrefix merges an environment variable alias with the configured prefix.
// The code bellow is from Viper's source.
func (c *Configurator) mergeWithEnvPrefix(in string) string {
	if c.envPrefix != "" {
		return strings.ToUpper(c.envPrefix + "_" + in)
	}

	return strings.ToUpper(in)
}

func (c *Configurator) Load(config interface{}) error {
	// Initial checks to see whether the config can be used as a target
	ptr := reflect.ValueOf(config)

	if ptr.Kind() != reflect.Ptr {
		return ErrNotStructPointer
	}

	elem := ptr.Elem()

	if elem.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	structType := elem.Type()

	if c.name == "" {
		c.name = c.args[0]
	}

	flags := pflag.NewFlagSet(c.name, pflag.ContinueOnError)
	var parseFlags bool

	// Gather configuration definition information
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)

		// Ignore unexported field
		if ast.IsExported(structField.Name) == false {
			continue
		}

		// Manually ignored field
		if value, ok := structField.Tag.Lookup("ignored"); ok && isTrue(value) {
			continue
		}

		field := elem.Field(i)

		// Resolve pointer to it's actual type
		for field.Kind() == reflect.Ptr {
			// Set to zero value when field is nil
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}

			field = field.Elem()
		}

		// Ignore unsupported field
		if _, unsupported := unsupportedTypes[field.Kind()]; unsupported {
			continue
		}

		// Set value override
		if value := field.Interface(); isZeroValueOfType(value) == false {
			c.viper.Set(structField.Name, value)
		}

		// Map flag to field
		if value, ok := structField.Tag.Lookup("flag"); ok {
			parseFlags = true

			// Use the field name as flag name if it is not provided
			if value == "" {
				// Make the first character lower case, because that's customary
				value = lowerFirst(structField.Name)

				// Try to split words in the struct name if possible
				if v, ok := structField.Tag.Lookup("split_words"); ok && isTrue(v) {
					v = splitWords(value, "-")
					if v != "" {
						value = v
					}
				}
			}

			// TODO: put default value here?
			flags.String(value, "", structField.Tag.Get("usage"))
			flag := flags.Lookup(value)

			c.viper.BindPFlag(structField.Name, flag)
		}

		// Map environment variable to field
		if value, ok := structField.Tag.Lookup("env"); ok {
			args := []string{structField.Name}

			// An environment variable alias is provided
			if value != "" {
				args = append(args, c.mergeWithEnvPrefix(value))
			} else if v, ok := structField.Tag.Lookup("split_words"); ok && isTrue(v) { // Try to split words in the struct name if possible
				v = splitWords(structField.Name, "_")
				if v != "" {
					args = append(args, c.mergeWithEnvPrefix(v))
				}
			}

			c.viper.BindEnv(args...)
		}

		// Set default (if any)
		if value, ok := structField.Tag.Lookup("default"); ok {
			c.viper.SetDefault(structField.Name, value)
		}
	}

	// Only parse flags if there is any
	if parseFlags {
		err := flags.Parse(c.args)
		if err == pflag.ErrHelp {
			return ErrFlagHelp
		} else if err != nil {
			return err
		}
	}

	// Apply configuration values
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)

		// Manually ignored field
		if value, ok := structField.Tag.Lookup("ignored"); ok && isTrue(value) {
			continue
		}

		// Check if value is present in Viper
		if c.viper.IsSet(structField.Name) == false {
			// Check for required value
			if value, ok := structField.Tag.Lookup("required"); ok && isTrue(value) {
				return fmt.Errorf("required field %s missing value", structField.Name)
			}

			// Ignore unset value
			continue
		}

		// Get the value from Viper
		value := c.viper.Get(structField.Name)

		field := elem.Field(i)

		if value != nil {
			// Process the value as string
			err := processField(field, fmt.Sprintf("%v", value))

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func processField(field reflect.Value, value string) error {
	typ := field.Type()

	// Resolve pointer to actual field and type
	// Zero value is already created earlier (when necessary)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		field = field.Elem()
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		var (
			val int64
			err error
		)

		if field.Kind() == reflect.Int64 && typ.PkgPath() == "time" && typ.Name() == "Duration" {
			var d time.Duration
			d, err = time.ParseDuration(value)
			val = int64(d)
		} else {
			val, err = strconv.ParseInt(value, 0, typ.Bits())
		}

		if err != nil {
			return err
		}

		field.SetInt(val)

	case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(value, 0, typ.Bits())
		if err != nil {
			return err
		}

		field.SetUint(val)

	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(value, typ.Bits())
		if err != nil {
			return err
		}

		field.SetFloat(val)

	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}

		field.SetBool(val)
	}

	return nil
}
