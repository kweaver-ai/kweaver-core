package cconf

type AuthzCfg struct {
	PrivateSvc *SvcConf `yaml:"private_svc"`
	PublicSvc  *SvcConf `yaml:"public_svc"`
}
