package configurator

import (
        "os"
	"github.com/pkg/errors"
)

// Configurator is configrator
type Configurator struct {
	loader     *loader
}

// Load is load
func (c *Configurator) Load(config interface{}) (error) {
	err := c.loader.load(config)
        return err
}

func validateConfigFile(configFile string) (error) {
        f, err := os.Open(configFile)
        if err != nil {
            return errors.Wrapf(err, "can not open config file (%v)", configFile)
        }
        f.Close()
        return nil
}

// NewConfigurator is create new configurator
func NewConfigurator(configFile string) (*Configurator, error) {
	err := validateConfigFile(configFile)
	if (err != nil) {
		return nil, errors.Wrapf(err, "invalid config file (%v)", configFile)
	}
	newConfigurator := &Configurator{
             loader: newLoader(configFile),
	}
	return newConfigurator, nil
}
