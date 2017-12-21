package nest

import (
	"errors"
	"fmt"
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

func NewConfigurator() *Configurator {
	return &Configurator{
		args:  os.Args,
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

	if c.name == "" {
		c.name = c.args[0]
	}

	flags := pflag.NewFlagSet(c.name, pflag.ContinueOnError)
	var parseFlags bool

	definitions := getDefinitions(elem)

	// Load definitions into Viper
	for _, def := range definitions {
		// Set value override
		if def.hasOverride {
			c.viper.Set(def.key, def.overrideValue)
		}

		// Map flag to field
		if def.hasFlag {
			parseFlags = true

			// TODO: put default value here?
			flags.String(def.flagAlias, "", def.usage)
			flag := flags.Lookup(def.flagAlias)

			c.viper.BindPFlag(def.key, flag)
		}

		// Map environment variable to field
		if def.hasEnv {
			var args []string

			// An environment variable alias is provided
			if def.envAlias != "" {
				args = []string{def.key, c.mergeWithEnvPrefix(def.envAlias)}
			} else {
				args = []string{def.key, c.mergeWithEnvPrefix(strings.Replace(def.key, ".", "_", -1))}
			}

			c.viper.BindEnv(args...)
		}

		// Set default (if any)
		if def.hasDefault {
			c.viper.SetDefault(def.key, def.defaultValue)
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
	for _, def := range definitions {
		// Check if value is present in Viper
		if c.viper.IsSet(def.key) == false {
			// Check for required value
			if def.required {
				return fmt.Errorf("required field %s missing value", def.key)
			}

			// Ignore unset value
			continue
		}

		// Get the value from Viper
		value := c.viper.Get(def.key)

		if value != nil {
			// Process the value as string
			err := processField(def.field, fmt.Sprintf("%v", value))

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func processField(field reflect.Value, value string) error {
	typ := field.Type()

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
