package oteltrace

import (
	"go.opentelemetry.io/otel/attribute"
)

func attributeValue(attrs []attribute.KeyValue, key string) string {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return attr.Value.AsString()
		}
	}

	return ""
}
