package nest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

	// ErrFlagHelp is returned when the commandline arguments include -h or --help.
	// Application should exit without an error as pflag handles outputting the manual.
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

	viper  *viper.Viper
	output io.Writer

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

// SetOutput sets the output writer used for help text and error messages.
func (c *Configurator) SetOutput(output io.Writer) {
	c.output = output
}

// out returns the configured output or the default which is STDERR.
func (c *Configurator) out() io.Writer {
	if c.output == nil {
		c.output = os.Stderr
	}

	return c.output
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
	flags.SetOutput(c.out())

	var parseFlags bool

	definitions := getDefinitions(elem)

	flags.Usage = func() {
		usage := getUsage(definitions)
		fmt.Fprintf(c.out(), "Usage of %s:\n", c.name)
		fmt.Fprint(c.out(), usage)
	}

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
			c.viper.BindEnv(def.key, c.mergeWithEnvPrefix(def.envAlias))
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
			// Format the value as string
			value := fmt.Sprintf("%v", value)

			// If the value is empty string, fall back to the zero value of the type
			if value == "" {
				value = fmt.Sprintf("%v", reflect.Zero(def.field.Type()).Interface())
			}

			// Process the value as string
			err := processField(def.field, value)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// getUsage returns the usage string for flags and environment variables.
func getUsage(definitions []fieldDefinition) string {
	buf := new(bytes.Buffer)

	var flagLines []string
	var envLines []string

	flagMaxlen := 0
	envMaxlen := 0

	for _, definition := range definitions {
		// Default value hint
		def := ""
		if definition.hasDefault {
			if definition.field.Type().Name() == "string" {
				def += fmt.Sprintf(" (default %q)", definition.defaultValue)
			} else {
				def += fmt.Sprintf(" (default %s)", definition.defaultValue)
			}
		}

		if definition.hasFlag {
			line := ""

			line = fmt.Sprintf("      --%s", definition.flagAlias)

			// Make an educated guess about the flag
			// TODO: check pflag UnquoteUsage
			name := definition.field.Type().Name()
			switch name {
			case "bool":
				name = ""
			case "float64":
				name = "float"
			case "int64":
				name = "int"
			case "uint64":
				name = "uint"
			}

			if name != "" {
				line += " " + name
			}

			// This special character will be replaced with spacing once the
			// correct alignment is calculated
			line += "\x00"
			if len(line) > flagMaxlen {
				flagMaxlen = len(line)
			}

			line += definition.usage
			line += def

			flagLines = append(flagLines, line)
		}

		if definition.hasEnv {
			line := ""

			line = fmt.Sprintf("      %s", c.mergeWithEnvPrefix(definition.envAlias))

			name := definition.field.Type().Name()
			switch name {
			case "float64":
				name = "float"
			case "int64":
				name = "int"
			case "uint64":
				name = "uint"
			}

			if name != "" {
				line += " " + name
			}

			// This special character will be replaced with spacing once the
			// correct alignment is calculated
			line += "\x00"
			if len(line) > envMaxlen {
				envMaxlen = len(line)
			}

			line += definition.usage
			line += def

			envLines = append(envLines, line)
		}
	}

	if len(flagLines) > 0 {
		fmt.Fprintln(buf, "\n\nFLAGS:\n")

		for _, line := range flagLines {
			sidx := strings.Index(line, "\x00")
			spacing := strings.Repeat(" ", flagMaxlen-sidx)
			// maxlen + 2 comes from + 1 for the \x00 and + 1 for the (deliberate) off-by-one in maxlen-sidx
			fmt.Fprintln(buf, line[:sidx], spacing, line[sidx+1:])
		}
	}

	if len(envLines) > 0 {
		fmt.Fprintln(buf, "\n\nENVIRONMENT VARIABLES:\n")

		for _, line := range envLines {
			sidx := strings.Index(line, "\x00")
			spacing := strings.Repeat(" ", envMaxlen-sidx)
			// maxlen + 2 comes from + 1 for the \x00 and + 1 for the (deliberate) off-by-one in maxlen-sidx
			fmt.Fprintln(buf, line[:sidx], spacing, line[sidx+1:])
		}
	}

	return buf.String()
}

func processField(field reflect.Value, value string) error {
	if canDecode(field) {
		return decode(field, value)
	}

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
