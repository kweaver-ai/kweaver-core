package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Mock responses for different services
type AppRegistrationResponse struct {
	ID        string `json:"id"`
	AppSecret string `json:"app_secret"`
	Name      string `json:"name"`
}

type ClusterAccessResponse struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	// User Management Service Mock (port 30980)
	go startUserManagementMock()

	// Deploy Service Mock (port 9703)
	go startDeployServiceMock()

	// Authentication Service Mock (port 30110)
	go startAuthenticationMock()

	// Hydra Admin Service Mock (port 4445)
	go startHydraMock()

	// Hydra Public Service Mock (port 4444) - OAuth2 token endpoint
	go startHydraPublicMock()

	// Business System Service Mock (port 8085)
	go startBusinessSystemMock()

	// Authorization Service Mock (port 30920)
	go startAuthorizationMock()

	// OSS Gateway Service Mock (port 9002)
	go startOSSGatewayMock()

	// Sandbox Service Mock (port 31700)
	go startSandboxMockServer()

	fmt.Println("Mock services started:")
	fmt.Println("  - User Management Service: http://localhost:30980")
	fmt.Println("  - Deploy Service: http://localhost:9703")
	fmt.Println("  - Authentication Service: http://localhost:30110")
	fmt.Println("  - Hydra Admin Service: http://localhost:4445")
	fmt.Println("  - Hydra Public Service: http://localhost:4444")
	fmt.Println("  - Business System Service: http://localhost:8085")
	fmt.Println("  - Authorization Service: http://localhost:30920")
	fmt.Println("  - OSS Gateway Service: http://localhost:9002")
	fmt.Println("Press Ctrl+C to stop...")

	select {}
}

func startUserManagementMock() {
	router := gin.New()
	router.Use(gin.Recovery())

	// Mock app registration endpoint
	router.POST("/api/user-management/v1/apps", func(c *gin.Context) {
		var requestBody map[string]interface{}
		if err := c.BindJSON(&requestBody); err != nil {
			log.Printf("[UserManagement] Invalid request body: %v", err)
		} else {
			log.Printf("[UserManagement] Received app registration request: %+v", requestBody)
		}

		response := AppRegistrationResponse{
			ID:        "mock-app-id-12345",
			AppSecret: "mock-app-secret-67890",
			Name:      "content-automation",
		}

		log.Printf("[UserManagement] Returning mock response: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	router.GET("/api/user-management/v1/users/:user_id/accessor_ids", func(c *gin.Context) {
		userID := c.Param("user_id")
		log.Printf("[UserManagement] Received accessor IDs request for user_id: %s", userID)

		response := []string{userID}

		log.Printf("[UserManagement] Returning mock response: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Mock get user info endpoint
	router.GET("/api/user-management/v1/users/:user_id/:fields", func(c *gin.Context) {
		userID := c.Param("user_id")
		fields := c.Param("fields")
		log.Printf("[UserManagement] Received get user info request for user_id: %s, fields: %s", userID, fields)

		// Return an array with one user object
		response := []gin.H{
			{
				"id":          userID,
				"name":        "mock-user",
				"csf_level":   1,
				"telephone":   "12345678901",
				"email":       "mock@example.com",
				"enabled":     true,
				"parent_deps": []interface{}{},
				"roles":       []string{"admin"},
				"custom_attr": gin.H{"is_knowledge": "1"},
			},
		}

		log.Printf("[UserManagement] Returning mock response: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Mock query internal account endpoint
	router.GET("/api/user-management/v1/apps/:app_id", func(c *gin.Context) {
		appID := c.Param("app_id")
		log.Printf("[UserManagement] Received query internal account request for app_id: %s", appID)

		response := AppRegistrationResponse{
			ID:        appID,
			AppSecret: "mock-app-secret-67890",
			Name:      "content-automation",
		}

		log.Printf("[UserManagement] Returning mock response: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "user-management-mock"})
	})

	port := os.Getenv("USER_MANAGEMENT_MOCK_PORT")
	if port == "" {
		port = "30980"
	}

	log.Printf("[UserManagement] Starting mock server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[UserManagement] Failed to start: %v", err)
	}
}

func startDeployServiceMock() {
	router := gin.New()
	router.Use(gin.Recovery())

	// Mock cluster access endpoint
	router.GET("/api/deploy-manager/v1/access-addr/app", func(c *gin.Context) {
		log.Printf("[DeployService] Received cluster access request")

		response := ClusterAccessResponse{
			Host: "localhost",
			Port: "8082",
		}

		log.Printf("[DeployService] Returning mock response: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "deploy-service-mock"})
	})

	port := os.Getenv("DEPLOY_SERVICE_MOCK_PORT")
	if port == "" {
		port = "9703"
	}

	log.Printf("[DeployService] Starting mock server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[DeployService] Failed to start: %v", err)
	}
}

func prettyJSON(data interface{}) string {
	bytes, _ := json.MarshalIndent(data, "", "  ")
	return string(bytes)
}

func startAuthenticationMock() {
	router := gin.New()
	router.Use(gin.Recovery())

	// Mock access token permission configuration endpoint
	router.PUT("/api/authentication/v1/access-token-perm/app/:app_id", func(c *gin.Context) {
		appID := c.Param("app_id")
		log.Printf("[Authentication] Received access token perm config request for app: %s", appID)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Mock JWT assertion endpoint
	router.GET("/api/authentication/v1/jwt", func(c *gin.Context) {
		userID := c.Query("user_id")
		log.Printf("[Authentication] Received JWT assertion request for user_id: %s", userID)

		// Return a mock JWT token
		mockJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6Ik1vY2sgVXNlciIsImlhdCI6MTUxNjIzOTAyMn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

		c.JSON(http.StatusOK, gin.H{
			"jwt": mockJWT,
		})
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "authentication-mock"})
	})

	port := os.Getenv("AUTHENTICATION_MOCK_PORT")
	if port == "" {
		port = "30110"
	}

	log.Printf("[Authentication] Starting mock server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[Authentication] Failed to start: %v", err)
	}
}

func startHydraMock() {
	router := gin.New()
	router.Use(gin.Recovery())

	// Mock introspection endpoint
	router.POST("/admin/oauth2/introspect", func(c *gin.Context) {
		log.Printf("[Hydra Admin] Received introspection request")
		c.JSON(http.StatusOK, gin.H{
			"active":    true,
			"scope":     "offline openid",
			"client_id": "mock-client-id",
			"sub":       "mock-user-id",
			"exp":       1735689600,
			"iat":       1704067200,
		})
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "hydra-admin-mock"})
	})

	port := os.Getenv("HYDRA_MOCK_PORT")
	if port == "" {
		port = "4445"
	}

	log.Printf("[Hydra Admin] Starting mock server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[Hydra Admin] Failed to start: %v", err)
	}
}

func startHydraPublicMock() {
	router := gin.New()
	router.Use(gin.Recovery())

	// Mock OAuth2 token endpoint - handles client_credentials grant
	router.POST("/oauth2/token", func(c *gin.Context) {
		// Parse form data (OAuth2 token requests use application/x-www-form-urlencoded)
		if err := c.Request.ParseForm(); err != nil {
			log.Printf("[Hydra Public] Failed to parse form data: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		log.Printf("[Hydra Public] Received token request (form): %+v", c.Request.PostForm)

		// Return a mock access token response
		response := gin.H{
			"access_token": "mock-access-token-" + fmt.Sprintf("%d", 1704067200),
			"token_type":   "bearer",
			"expires_in":   3600,
			"scope":        "offline openid",
		}

		log.Printf("[Hydra Public] Returning mock token response: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "hydra-public-mock"})
	})

	port := os.Getenv("HYDRA_PUBLIC_MOCK_PORT")
	if port == "" {
		port = "4444"
	}

	log.Printf("[Hydra Public] Starting mock server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[Hydra Public] Failed to start: %v", err)
	}
}

func startBusinessSystemMock() {
	router := gin.New()
	router.Use(gin.Recovery())

	// Mock resource binding endpoint (POST)
	router.POST("/internal/api/business-system/v1/resource", func(c *gin.Context) {
		var requestBody map[string]interface{}
		if err := c.BindJSON(&requestBody); err != nil {
			log.Printf("[BusinessSystem] Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		log.Printf("[BusinessSystem] Received resource binding request: %+v", requestBody)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Mock resource unbinding endpoint (DELETE)
	router.DELETE("/internal/api/business-system/v1/resource", func(c *gin.Context) {
		bdID := c.Query("bd_id")
		id := c.Query("id")
		dtype := c.Query("type")
		log.Printf("[BusinessSystem] Received resource unbinding request: bd_id=%s, id=%s, type=%s", bdID, id, dtype)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Mock resource list endpoint (GET)
	router.GET("/api/business-system/v1/resource", func(c *gin.Context) {
		bdID := c.Query("bd_id")
		limit := c.Query("limit")
		offset := c.Query("offset")
		log.Printf("[BusinessSystem] Received resource list request: bd_id=%s, limit=%s, offset=%s", bdID, limit, offset)

		response := gin.H{
			"total": 0,
			"items": []interface{}{},
		}

		c.JSON(http.StatusOK, response)
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "business-system-mock"})
	})

	port := os.Getenv("BUSINESS_SYSTEM_MOCK_PORT")
	if port == "" {
		port = "8085"
	}

	log.Printf("[BusinessSystem] Starting mock server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[BusinessSystem] Failed to start: %v", err)
	}
}

func startAuthorizationMock() {
	router := gin.New()
	router.Use(gin.Recovery())

	// Mock operation permission check endpoint
	router.POST("/api/authorization/v1/operation-check", func(c *gin.Context) {
		var requestBody map[string]interface{}
		if err := c.BindJSON(&requestBody); err != nil {
			log.Printf("[Authorization] Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		log.Printf("[Authorization] Received operation-check request: %+v", requestBody)

		// Return success with result: true (permission granted)
		c.JSON(http.StatusOK, gin.H{
			"result": true,
		})
	})

	// Mock create policy endpoint
	router.POST("/api/authorization/v1/policy", func(c *gin.Context) {
		var requestBody []map[string]interface{}
		if err := c.BindJSON(&requestBody); err != nil {
			log.Printf("[Authorization] Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		log.Printf("[Authorization] Received create policy request: %+v", requestBody)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Mock update policy endpoint
	router.PUT("/api/authorization/v1/policy/:policy_ids", func(c *gin.Context) {
		policyIDs := c.Param("policy_ids")
		var requestBody []map[string]interface{}
		if err := c.BindJSON(&requestBody); err != nil {
			log.Printf("[Authorization] Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		log.Printf("[Authorization] Received update policy request for IDs %s: %+v", policyIDs, requestBody)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Mock delete policy endpoint
	router.POST("/api/authorization/v1/policy-delete", func(c *gin.Context) {
		var requestBody map[string]interface{}
		if err := c.BindJSON(&requestBody); err != nil {
			log.Printf("[Authorization] Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		log.Printf("[Authorization] Received delete policy request: %+v", requestBody)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Mock resource filter endpoint
	router.POST("/api/authorization/v1/resource-filter", func(c *gin.Context) {
		var requestBody map[string]interface{}
		if err := c.BindJSON(&requestBody); err != nil {
			log.Printf("[Authorization] Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		log.Printf("[Authorization] Received resource-filter request: %+v", requestBody)

		// Return the same resources that were sent (all pass the filter)
		resources := requestBody["resources"]
		if resources == nil {
			resources = []interface{}{}
		}
		c.JSON(http.StatusOK, resources)
	})

	// Mock resource operation list endpoint
	router.POST("/api/authorization/v1/resource-operation", func(c *gin.Context) {
		var requestBody map[string]interface{}
		if err := c.BindJSON(&requestBody); err != nil {
			log.Printf("[Authorization] Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		log.Printf("[Authorization] Received resource-operation request: %+v", requestBody)

		// Return mock operation list for each resource
		resources := []gin.H{}
		if reqResources, ok := requestBody["resources"].([]interface{}); ok {
			for _, res := range reqResources {
				if resMap, ok := res.(map[string]interface{}); ok {
					resources = append(resources, gin.H{
						"id":        resMap["id"],
						"operation": []string{"read", "write", "delete", "execute"},
					})
				}
			}
		}
		c.JSON(http.StatusOK, resources)
	})

	// Mock resource list endpoint
	router.POST("/api/authorization/v1/resource-list", func(c *gin.Context) {
		var requestBody map[string]interface{}
		if err := c.BindJSON(&requestBody); err != nil {
			log.Printf("[Authorization] Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		log.Printf("[Authorization] Received resource-list request: %+v", requestBody)

		// Return empty resource list
		c.JSON(http.StatusOK, []interface{}{})
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "authorization-mock"})
	})

	port := os.Getenv("AUTHORIZATION_MOCK_PORT")
	if port == "" {
		port = "30920"
	}

	log.Printf("[Authorization] Starting mock server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[Authorization] Failed to start: %v", err)
	}
}

func startOSSGatewayMock() {
	router := gin.New()
	router.Use(gin.Recovery())

	// Mock get default storage endpoint
	router.GET("/api/ossgateway/v1/default-storage", func(c *gin.Context) {
		log.Printf("[OSSGateway] Received get default storage request")

		response := gin.H{
			"storage_id": "mock-oss-storage-id",
			"name":       "Mock OSS Storage",
			"enabled":    true,
			"default":    true,
		}

		log.Printf("[OSSGateway] Returning mock response: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Mock get object storage info endpoint
	router.GET("/api/ossgateway/v1/objectstorageinfo", func(c *gin.Context) {
		app := c.Query("app")
		log.Printf("[OSSGateway] Received get object storage info request for app: %s", app)

		response := []gin.H{
			{
				"id":      "mock-oss-storage-id",
				"name":    "Mock OSS Storage",
				"enabled": true,
				"default": true,
			},
		}

		log.Printf("[OSSGateway] Returning mock response: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Mock upload endpoint - returns request body for actual upload
	router.GET("/api/ossgateway/v1/upload/:oss_id/*key", func(c *gin.Context) {
		ossID := c.Param("oss_id")
		key := c.Param("key")
		requestMethod := c.Query("request_method")
		internalRequest := c.Query("internal_request")
		log.Printf("[OSSGateway] Received upload request for oss_id: %s, key: %s, method: %s, internal: %s",
			ossID, key, requestMethod, internalRequest)

		// Return a mock request body that simulates what the real OSS gateway would return
		response := gin.H{
			"method": "PUT",
			"url":    "http://localhost:9002/mock-oss-storage/upload",
			"headers": gin.H{
				"Content-Type": "application/octet-stream",
			},
		}

		log.Printf("[OSSGateway] Returning mock upload request body: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Mock actual upload endpoint (receives the file data)
	router.PUT("/mock-oss-storage/upload", func(c *gin.Context) {
		log.Printf("[OSSGateway] Received actual file upload")
		c.Status(http.StatusOK)
	})

	// Mock download endpoint
	router.GET("/api/ossgateway/v1/download/:oss_id/*key", func(c *gin.Context) {
		ossID := c.Param("oss_id")
		key := c.Param("key")
		internalRequest := c.Query("internal_request")
		downloadType := c.Query("type")
		log.Printf("[OSSGateway] Received download request for oss_id: %s, key: %s, internal: %s, type: %s",
			ossID, key, internalRequest, downloadType)

		response := gin.H{
			"method": "GET",
			"url":    "http://localhost:9002/mock-oss-storage/download",
			"headers": gin.H{
				"Content-Type": "application/octet-stream",
			},
		}

		log.Printf("[OSSGateway] Returning mock download request body: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Mock actual download endpoint
	router.GET("/mock-oss-storage/download", func(c *gin.Context) {
		log.Printf("[OSSGateway] Received actual file download")
		c.JSON(http.StatusOK, gin.H{"data": "mock file content"})
	})

	// Mock delete endpoint
	router.GET("/api/ossgateway/v1/delete/:oss_id/*key", func(c *gin.Context) {
		ossID := c.Param("oss_id")
		key := c.Param("key")
		internalRequest := c.Query("internal_request")
		log.Printf("[OSSGateway] Received delete request for oss_id: %s, key: %s, internal: %s",
			ossID, key, internalRequest)

		response := gin.H{
			"method":  "DELETE",
			"url":     "http://localhost:9002/mock-oss-storage/delete",
			"headers": gin.H{},
		}

		log.Printf("[OSSGateway] Returning mock delete request body: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Mock actual delete endpoint
	router.DELETE("/mock-oss-storage/delete", func(c *gin.Context) {
		log.Printf("[OSSGateway] Received actual file delete")
		c.Status(http.StatusOK)
	})

	// Mock head (metadata) endpoint
	router.GET("/api/ossgateway/v1/head/:oss_id/*key", func(c *gin.Context) {
		ossID := c.Param("oss_id")
		key := c.Param("key")
		internalRequest := c.Query("internal_request")
		log.Printf("[OSSGateway] Received head request for oss_id: %s, key: %s, internal: %s",
			ossID, key, internalRequest)

		response := gin.H{
			"method":  "HEAD",
			"url":     "http://localhost:9002/mock-oss-storage/head",
			"headers": gin.H{},
		}

		log.Printf("[OSSGateway] Returning mock head request body: %+v", response)
		c.JSON(http.StatusOK, response)
	})

	// Mock actual head endpoint
	router.HEAD("/mock-oss-storage/head", func(c *gin.Context) {
		log.Printf("[OSSGateway] Received actual file head request")
		c.Header("Content-Length", "1024")
		c.Status(http.StatusOK)
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "ossgateway-mock"})
	})

	port := os.Getenv("OSSGATEWAY_MOCK_PORT")
	if port == "" {
		port = "9002"
	}

	log.Printf("[OSSGateway] Starting mock server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[OSSGateway] Failed to start: %v", err)
	}
}
