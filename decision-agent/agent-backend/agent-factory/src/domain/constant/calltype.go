package constant

type CallType string

const (
	Chat         CallType = "chat"
	DebugChat    CallType = "debug_chat"
	APIChat      CallType = "api_chat"
	InternalChat CallType = "internal_chat"
)

func (c CallType) String() string {
	return string(c)
}
