package contextcreator

type Config struct {
	configs map[string]ContextConfig
}

func (c *Config) Set(providerID string, config ContextConfig) {
	c.configs[providerID] = config
}

func (c *Config) Get(providerID string) ContextConfig {
	config, _ := c.configs[providerID]
	return config
}

type ContextConfig struct {
	TLSSkipVerify bool
}
