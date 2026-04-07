package icmp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockMQClient is a mock implementation of IMQClient for testing
type MockMQClient struct {
	PublishCalled    bool
	SubscribeCalled  bool
	CloseCalled      bool
	PublishTopic     string
	PublishData      []byte
	SubscribeTopic   string
	SubscribeChannel string
}

func (m *MockMQClient) Publish(topic string, data []byte) error {
	m.PublishCalled = true
	m.PublishTopic = topic
	m.PublishData = data

	return nil
}

func (m *MockMQClient) Subscribe(topic, channel string, cmd func([]byte) error) error {
	m.SubscribeCalled = true
	m.SubscribeTopic = topic
	m.SubscribeChannel = channel

	return nil
}

func (m *MockMQClient) Close() {
	m.CloseCalled = true
}

func TestIMQClient_Interface(t *testing.T) {
	t.Parallel()

	// Test that MockMQClient implements IMQClient
	var _ IMQClient = &MockMQClient{}

	assert.True(t, true)
}

func TestMockMQClient_Publish(t *testing.T) {
	t.Parallel()

	mock := &MockMQClient{}
	data := []byte("test data")

	err := mock.Publish("test.topic", data)

	assert.NoError(t, err)
	assert.True(t, mock.PublishCalled)
	assert.Equal(t, "test.topic", mock.PublishTopic)
	assert.Equal(t, data, mock.PublishData)
}

func TestMockMQClient_Subscribe(t *testing.T) {
	t.Parallel()

	mock := &MockMQClient{}
	cmd := func(data []byte) error { return nil }

	err := mock.Subscribe("test.topic", "test.channel", cmd)

	assert.NoError(t, err)
	assert.True(t, mock.SubscribeCalled)
	assert.Equal(t, "test.topic", mock.SubscribeTopic)
	assert.Equal(t, "test.channel", mock.SubscribeChannel)
}

func TestMockMQClient_Close(t *testing.T) {
	t.Parallel()

	mock := &MockMQClient{}

	mock.Close()

	assert.True(t, mock.CloseCalled)
}
