package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	*viper.Viper
}

var configFile string = "config.yaml"

// NewNakedConfig creates a new config without initializing it
// useful for testing
func NewNakedConfig() *Config {
	conf := &Config{
		viper.New(),
	}
	return conf
}

// NewConfig creates a new config and initializes it
func NewConfig() (*Config, error) {
	conf := &Config{
		viper.New(),
	}
	err := conf.readFullConfig()
	return conf, err
}

// readFullConfig reads the full config from the config file and environment variables
func (conf *Config) readFullConfig() error {
	for key, configVar := range configVars {
		conf.SetDefault(key, configVar.Default)
		if configVar.EnvVar != "" {
			conf.BindEnv(key, configVar.EnvVar)
		}
	}
	conf.SetConfigName(configFile)
	conf.SetConfigType("yaml")
	conf.AddConfigPath(".")

	if err := conf.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return err
		} else {
			log.Fatalf("error: %+v", err)
		}
	}
	return nil
}
