package conf

type HTTPServerConfig struct {
	Address string
}

func NewHTTPServerConfig() HTTPServerConfig {
	return HTTPServerConfig{
		Address: ":8080",
	}
}
