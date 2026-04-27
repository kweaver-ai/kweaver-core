package cconf

type BizDomainConf struct {
	PrivateSvc *BizDomainSvcConf `yaml:"private_svc"`
}

type BizDomainSvcConf struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol"`
}
