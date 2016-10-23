package contextcacher

type Config struct {
	cacheContextSize        int
	cachePrivateContextSize int
	configs                 map[string]ContextConfig
}

func NewConfig(cacheContextSize, cachePrivateContextSize int) *Config {
	return &Config{
		cacheContextSize:        cacheContextSize,
		cachePrivateContextSize: cachePrivateContextSize,
		configs:                 make(map[string]ContextConfig),
	}
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
