package apidocs

import _ "embed"

var (
	//go:embed agent-factory.json
	AgentFactoryJSON []byte

	//go:embed agent-factory.yaml
	AgentFactoryYAML []byte

	//go:embed agent-factory.html
	AgentFactoryHTML []byte
)
