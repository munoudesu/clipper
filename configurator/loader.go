package configurator

import (
	"github.com/pkg/errors"
	"github.com/BurntSushi/toml"
)

type loader struct {
	configFile string
}

func (l *loader) loadToml(config interface{}) (error) {
	_, err := toml.DecodeFile(l.configFile, config)
	if err != nil {
		return errors.Wrapf(err, "can not decode config file for toml (%v)", l.configFile)
	}
	return nil
}

func (l *loader)load(config interface{}) (error) {
	err := l.loadToml(config)
	if err != nil {
		return errors.Wrapf(err, "can not load config file (%v)", l.configFile)
	}
	return nil
}

func newLoader(configFile string) (*loader) {
	return &loader{
            configFile: configFile,
        }
}
