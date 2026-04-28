package config

type ModuleConfig struct {
}

type SatusehatPropagationConfig struct {
}

func (c *ModuleConfig) SetEnvBindings() map[string]string {
	return map[string]string{}
}

func (c *ModuleConfig) SetDefaults() map[string]any {
	return map[string]any{}
}
