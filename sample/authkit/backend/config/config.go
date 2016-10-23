package config

import (
	"fmt"
	"log"
	"time"

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
	err = viper.Unmarshal(&c.c)
	if err != nil {
		panic(err)
	}
	c.c.PrivateProviderID = "hydra-sample"
	c.c.PrivateProviderIDTrustedContext = "hydra-sample-trusted"
	c.c.PrivateOAuth2Provider.ID = c.c.PrivateProviderID
	c.c.modTime = time.Now()

	log.Printf("Effective config:\n%#v\n", c.c)
}

func Get() Config {
	return c
}
