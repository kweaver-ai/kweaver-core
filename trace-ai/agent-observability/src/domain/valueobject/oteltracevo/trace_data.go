package oteltracevo

type TraceData struct {
	ResourceSpans []ResourceSpan `json:"resourceSpans"`
}

type ResourceSpan struct {
	Resource   Resource    `json:"resource"`
	ScopeSpans []ScopeSpan `json:"scopeSpans"`
}

type Resource struct {
	Attributes []Attribute `json:"attributes"`
}

type ScopeSpan struct {
	Scope Scope  `json:"scope"`
	Spans []Span `json:"spans"`
}

type Scope struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type Span struct {
	TraceID           string      `json:"traceId"`
	SpanID            string      `json:"spanId"`
	ParentSpanID      string      `json:"parentSpanId,omitempty"`
	Name              string      `json:"name"`
	Kind              string      `json:"kind"`
	StartTimeUnixNano string      `json:"startTimeUnixNano"`
	EndTimeUnixNano   string      `json:"endTimeUnixNano"`
	Attributes        []Attribute `json:"attributes,omitempty"`
	Events            []Event     `json:"events,omitempty"`
	Status            Status      `json:"status"`
}

type Attribute struct {
	Key   string         `json:"key"`
	Value AttributeValue `json:"value"`
}

type AttributeValue struct {
	StringValue string  `json:"stringValue,omitempty"`
	IntValue    string  `json:"intValue,omitempty"`
	BoolValue   bool    `json:"boolValue,omitempty"`
	DoubleValue float64 `json:"doubleValue,omitempty"`
}

type Event struct {
	Name         string      `json:"name"`
	TimeUnixNano string      `json:"timeUnixNano"`
	Attributes   []Attribute `json:"attributes,omitempty"`
}

type Status struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
