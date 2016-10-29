package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Init initializes app's config.
// App should invoke this function from the main after it parsed flags.
// prefPath, prefName allows to overwrite default values for config dir and base file name.
func Init(prefPath, prefName string) {
	if prefPath != "" {
		viper.AddConfigPath(prefPath)
	}
	viper.AddConfigPath(defPath)
	viper.AddConfigPath(".")
	if prefName == "" {
		prefName = defName
	}
	viper.SetConfigName(prefName)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	c := &config{}
	err = viper.Unmarshal(c)
	if err != nil {
		panic(err)
	}
	c.init()
	cfg = Config{&configWrapper{c}}

	// log.Printf("Effective config:\n%#v\n", c.c)
}

func Get() Config {
	return cfg
}

var cfg Config
