package apidocs

import _ "embed"

var (
	//go:embed assets/agent-factory.json
	AgentFactoryJSON []byte

	//go:embed assets/agent-factory.yaml
	AgentFactoryYAML []byte

	//go:embed assets/agent-factory.html
	AgentFactoryHTML []byte

	//go:embed assets/favicon.png
	AgentFactoryFaviconPNG []byte
)
