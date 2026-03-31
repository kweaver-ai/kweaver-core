package ctopicenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPmsMqTopic_AgentNameModify(t *testing.T) {
	t.Parallel()

	assert.Equal(t, PmsMqTopic("authorization.resource.name.modify"), AgentNameModifyForAuthorizationPlatform)
}

func TestPmsMqTopic_NotEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, AgentNameModifyForAuthorizationPlatform)
}

func TestPmsMqTopic_String(t *testing.T) {
	t.Parallel()

	topic := PmsMqTopic(AgentNameModifyForAuthorizationPlatform)
	assert.Equal(t, "authorization.resource.name.modify", string(topic))
}

func TestPmsMqTopic_TypeAlias(t *testing.T) {
	t.Parallel()

	var topic PmsMqTopic = "test.topic"

	assert.Equal(t, MqTopic("test.topic"), topic)
}
