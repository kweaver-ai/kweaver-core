package cconf

// Hydra 相关配置
type HydraCfg struct {
	UserMgnt    UserMgntCfg    `yaml:"user-mgnt"`
	HydraPublic HydraPublicCfg `yaml:"hydra-public"`
	HydraAdmin  HydraAdminCfg  `yaml:"hydra-admin"`
}

type UserMgntCfg struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol"`
}

type HydraPublicCfg struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type HydraAdminCfg struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}
