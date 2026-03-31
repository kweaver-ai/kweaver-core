package cconf

type DBConf struct {
	UserName string `yaml:"user_name"`
	Password string `yaml:"user_pwd"`
	DBHost   string `yaml:"db_host"`
	DBPort   int    `yaml:"db_port"`
	DBName   string `yaml:"db_name"`

	Charset          string `yaml:"db_charset"`
	Timeout          int    `yaml:"timeout"`
	TimeoutRead      int    `yaml:"read_timeout"`
	TimeoutWrite     int    `yaml:"write_timeout"`
	MaxOpenConns     int    `yaml:"max_open_conns"`
	MaxOpenReadConns int    `yaml:"max_open_read_conns"`
}
