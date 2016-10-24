package contextcacher

type Config struct {
	cacheContextSize int
	configs          map[string]ContextConfig
}

func NewConfig(cacheContextSize int) *Config {
	return &Config{
		cacheContextSize: cacheContextSize,
		configs:          make(map[string]ContextConfig),
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
