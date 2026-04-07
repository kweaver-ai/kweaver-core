package cutil

const (
	// SandboxSessionID 统一的 Sandbox Session ID
	SandboxSessionID = "sess-agent-default"
)

// GetSandboxSessionID 获取 Sandbox Session ID
func GetSandboxSessionID() string {
	return SandboxSessionID
}
