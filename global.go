package nest

// c is a global Configurator instance following Viper's singleton principle.
var c *Configurator

func init() {
	c = NewConfigurator()
}

// SetEnvPrefix calls the function with the same name on the global configurator instance.
func SetEnvPrefix(prefix string) {
	c.SetEnvPrefix(prefix)
}

// SetName calls the function with the same name on the global configurator instance.
func SetName(name string) {
	c.SetName(name)
}

// Load calls the function with the same name on the global configurator instance.
func Load(config interface{}) error {
	return c.Load(config)
}
