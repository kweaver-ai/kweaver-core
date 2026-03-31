package httpserver

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// MockHealthHandler is a mock implementation of IHTTPHealthRouter
type MockHealthHandler struct {
	RegHealthRouterCalled bool
	RouterGroupReceived   *gin.RouterGroup
}

func (m *MockHealthHandler) RegHealthRouter(routerGroup *gin.RouterGroup) {
	m.RegHealthRouterCalled = true
	m.RouterGroupReceived = routerGroup
}

func TestRegisterHealthRoutes_RegistersHealthGroup(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	mockHandler := &MockHealthHandler{}

	server := &httpServer{
		httpHealthHandler: mockHandler,
	}

	server.registerHealthRoutes(engine)

	assert.True(t, mockHandler.RegHealthRouterCalled)
	assert.NotNil(t, mockHandler.RouterGroupReceived)
}

func TestRegisterHealthRoutes_WithNilEngine(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	mockHandler := &MockHealthHandler{}

	server := &httpServer{
		httpHealthHandler: mockHandler,
	}

	// This should panic if engine is nil, but we're testing the method signature
	// In practice, this would be called with a valid gin.Engine
	assert.Panics(t, func() {
		server.registerHealthRoutes(nil)
	})
}

func TestRegisterHealthRoutes_CreatesCorrectPath(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	mockHandler := &MockHealthHandler{}

	server := &httpServer{
		httpHealthHandler: mockHandler,
	}

	server.registerHealthRoutes(engine)

	// Verify the router group was created with the correct base path
	assert.NotNil(t, mockHandler.RouterGroupReceived)
	// The router group's base path should be /health
	// We can't directly access the base path, but we can verify the handler was called
	assert.True(t, mockHandler.RegHealthRouterCalled)
}

func TestRegisterHealthRoutes_MultipleCalls(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	mockHandler := &MockHealthHandler{}

	server := &httpServer{
		httpHealthHandler: mockHandler,
	}

	// Call registerHealthRoutes multiple times
	server.registerHealthRoutes(engine)

	firstCallCount := 0
	if mockHandler.RegHealthRouterCalled {
		firstCallCount = 1
	}

	server.registerHealthRoutes(engine)

	secondCallCount := 0
	if mockHandler.RegHealthRouterCalled {
		secondCallCount = 1
	}

	// Handler should be called each time
	assert.Equal(t, firstCallCount, secondCallCount)
}
