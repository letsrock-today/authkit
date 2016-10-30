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
	cfg = Config{}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}
	cfg.init()

	//log.Printf("Effective config:\n%#v\n" cfg, *cfg.PrivateOAuth2Provider)
}

func Get() Config {
	return cfg
}

var cfg Config
