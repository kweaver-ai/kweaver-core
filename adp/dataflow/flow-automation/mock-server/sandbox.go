package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func startSandboxMockServer() {
	router := gin.New()
	router.Use(gin.Recovery())
	// Mock create session endpoint
	router.POST("/api/v1/sessions", func(c *gin.Context) {
		var requestBody map[string]interface{}
		if err := c.BindJSON(&requestBody); err != nil {
			log.Printf("[Sandbox] Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		log.Printf("[Sandbox] Received create session request: %+v", requestBody)

		response := gin.H{
			"id":             "mock-session-" + fmt.Sprintf("%d", 1704067200),
			"template_id":    requestBody["template_id"],
			"status":         "running",
			"workspace_path": "/workspace/mock-session",
			"runtime_type":   "python",
			"container_id":   "mock-container-id",
			"timeout":        3600,
			"created_at":     "2024-01-01T00:00:00Z",
			"updated_at":     "2024-01-01T00:00:00Z",
		}

		log.Printf("[Sandbox] Returning mock session response: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Mock get session endpoint
	router.GET("/api/v1/sessions/:session_id", func(c *gin.Context) {
		sessionID := c.Param("session_id")
		log.Printf("[Sandbox] Received get session request for session_id: %s", sessionID)

		response := gin.H{
			"id":             sessionID,
			"template_id":    "mock-template-id",
			"status":         "running",
			"workspace_path": "/workspace/mock-session",
			"runtime_type":   "python",
			"container_id":   "mock-container-id",
			"timeout":        3600,
			"created_at":     "2024-01-01T00:00:00Z",
			"updated_at":     "2024-01-01T00:00:00Z",
		}

		log.Printf("[Sandbox] Returning mock session: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Mock execute code endpoint
	router.POST("/api/v1/executions/sessions/:session_id/execute", func(c *gin.Context) {
		sessionID := c.Param("session_id")
		var requestBody map[string]interface{}
		if err := c.BindJSON(&requestBody); err != nil {
			log.Printf("[Sandbox] Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		log.Printf("[Sandbox] Received execute code request for session_id: %s, body: %+v", sessionID, requestBody)

		response := gin.H{
			"execution_id": "mock-execution-" + fmt.Sprintf("%d", 1704067200),
			"session_id":   sessionID,
			"status":       "running",
			"created_at":   "2024-01-01T00:00:00Z",
		}

		log.Printf("[Sandbox] Returning mock execution: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Mock get execution status endpoint
	router.GET("/api/v1/executions/:execution_id/status", func(c *gin.Context) {
		executionID := c.Param("execution_id")
		log.Printf("[Sandbox] Received get execution status request for execution_id: %s", executionID)

		response := gin.H{
			"id":             executionID,
			"session_id":     "mock-session-id",
			"status":         "completed",
			"code":           "print('hello')",
			"language":       "python",
			"timeout":        3600,
			"exit_code":      0,
			"execution_time": 0.5,
			"stdout":         "hello\n",
			"stderr":         "",
			"artifacts":      []interface{}{},
			"retry_count":    0,
			"created_at":     "2024-01-01T00:00:00Z",
			"started_at":     "2024-01-01T00:00:01Z",
			"completed_at":   "2024-01-01T00:00:02Z",
		}

		log.Printf("[Sandbox] Returning mock execution status: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Mock get execution result endpoint
	router.GET("/api/v1/executions/:execution_id/result", func(c *gin.Context) {
		executionID := c.Param("execution_id")
		log.Printf("[Sandbox] Received get execution result request for execution_id: %s", executionID)

		response := gin.H{
			"id":             executionID,
			"session_id":     "mock-session-id",
			"status":         "completed",
			"code":           "print('hello')",
			"language":       "python",
			"timeout":        3600,
			"exit_code":      0,
			"execution_time": 0.5,
			"stdout":         "hello\n",
			"stderr":         "",
			"artifacts":      []interface{}{},
			"retry_count":    0,
			"created_at":     "2024-01-01T00:00:00Z",
			"started_at":     "2024-01-01T00:00:01Z",
			"completed_at":   "2024-01-01T00:00:02Z",
			"return_value": gin.H{
				"result": "success",
			},
			"metrics": gin.H{
				"memory_used": "128MB",
				"cpu_used":    "10%",
			},
		}

		log.Printf("[Sandbox] Returning mock execution result: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "sandbox-mock"})
	})

	port := os.Getenv("SANDBOX_MOCK_PORT")
	if port == "" {
		port = "31700"
	}

	log.Printf("[Sandbox] Starting mock server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[Sandbox] Failed to start: %v", err)
	}

}
