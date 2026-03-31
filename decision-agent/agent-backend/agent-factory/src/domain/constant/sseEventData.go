package constant

import "fmt"

const (
	EventStreamEventEnd = "end"
	SSEFieldData        = "data"
)

var DataEventEndStr = fmt.Sprintf("%s: %s", SSEFieldData, "event: "+EventStreamEventEnd)
