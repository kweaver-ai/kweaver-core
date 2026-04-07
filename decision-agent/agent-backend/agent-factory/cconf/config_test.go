package cconf

import (
	"os"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

func TestMain(m *testing.M) {
	// CRITICAL: cenvhelper's TestMain sets SERVICE_NAME to "mock_svc_name"
	// We need to override it for cconf tests
	os.Setenv("SERVICE_NAME", "AGENT_FACTORY")

	// Reinitialize cenvhelper with the correct SERVICE_NAME
	cenvhelper.InitEnvForTest()

	// Run tests
	code := m.Run()

	os.Exit(code)
}

func TestGetConfigPath(t *testing.T) {
	t.Run("default config path", func(t *testing.T) {
		// Reset the global variable
		_configPath = ""

		// Save original env value
		originalPath := os.Getenv("CONFIG_PATH")

		// Clean up after test
		defer func() {
			_configPath = ""

			if originalPath != "" {
				os.Setenv("CONFIG_PATH", originalPath)
			} else {
				os.Unsetenv("CONFIG_PATH")
			}
		}()

		path := GetConfigPath()

		// The path will be either /sysvol/conf or ./conf depending on what exists
		if path == "" {
			t.Error("Expected GetConfigPath to return a non-empty string")
		}
	})
}

func TestConfig_IsDebug(t *testing.T) {
	t.Run("debug mode false by default", func(t *testing.T) {
		// Save original env value
		originalDebug := os.Getenv("DEBUG_MODE")

		// Clean up after test
		defer func() {
			if originalDebug != "" {
				os.Setenv("DEBUG_MODE", originalDebug)
			} else {
				os.Unsetenv("DEBUG_MODE")
			}
		}()

		// Ensure DEBUG_MODE is not set
		os.Unsetenv("DEBUG_MODE")

		config := &Config{
			Project: Project{
				Host: "localhost",
				Port: 8080,
			},
		}

		if config.IsDebug() {
			t.Error("Expected IsDebug to return false by default")
		}
	})
}

func TestConfig_GetDefaultLanguage(t *testing.T) {
	t.Run("simplified chinese", func(t *testing.T) {
		config := &Config{
			Project: Project{
				Language: rest.SimplifiedChinese,
			},
		}

		lang := config.GetDefaultLanguage()
		if lang != rest.SimplifiedChinese {
			t.Errorf("Expected SimplifiedChinese, got %v", lang)
		}
	})

	t.Run("american english", func(t *testing.T) {
		config := &Config{
			Project: Project{
				Language: rest.AmericanEnglish,
			},
		}

		lang := config.GetDefaultLanguage()
		if lang != rest.AmericanEnglish {
			t.Errorf("Expected AmericanEnglish, got %v", lang)
		}
	})
}

func TestConfig_GetLogLevelString(t *testing.T) {
	t.Run("log level 1", func(t *testing.T) {
		config := &Config{
			Project: Project{
				LoggerLevel: 1,
			},
		}

		level := config.GetLogLevelString()
		if level == "" {
			t.Error("Expected GetLogLevelString to return a non-empty string")
		}
	})
}

func TestConfig_String(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &Config{
			Project: Project{
				Host: "localhost",
				Port: 8080,
			},
		}

		str := config.String()
		if str == "" {
			t.Error("Expected String to return a non-empty string")
		}
	})
}

func TestBaseDefConfig(t *testing.T) {
	t.Run("default config values", func(t *testing.T) {
		config := BaseDefConfig()

		if config == nil {
			t.Fatal("Expected BaseDefConfig to return a non-nil config")
		}

		if config.Project.Host != "0.0.0.0" {
			t.Errorf("Expected Host to be '0.0.0.0', got '%s'", config.Project.Host)
		}

		if config.Project.Port != 30777 {
			t.Errorf("Expected Port to be 30777, got %d", config.Project.Port)
		}

		if config.Project.Language != rest.SimplifiedChinese {
			t.Errorf("Expected Language to be SimplifiedChinese, got %v", config.Project.Language)
		}
		// Note: Debug mode is controlled by environment variable, not Project struct
		// BaseDefConfig returns a config, debug state depends on environment
		_ = config.Project // Just verify the config exists
	})
}

func TestGetConfigBys(t *testing.T) {
	t.Run("test function exists", func(t *testing.T) {
		// This test verifies that GetConfigBys function exists
		// The actual functionality depends on file system
		// which is difficult to test in unit tests
		// Note: Calling this with a non-existent file will call log.Fatalf
		// which will terminate the test
		_ = GetConfigBys
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("test function exists", func(t *testing.T) {
		// This test verifies that LoadConfig function exists
		// The actual functionality depends on YAML unmarshaling
		// which is difficult to test in unit tests
		// Note: This will call log.Fatalf if the YAML is invalid
		_ = LoadConfig
	})
}

func TestConfig_Check(t *testing.T) {
	t.Run("valid project config", func(t *testing.T) {
		config := &Config{
			Project: Project{
				Host:     "localhost",
				Port:     8080,
				Language: rest.SimplifiedChinese,
			},
		}

		err := config.Check()
		if err != nil {
			t.Errorf("Expected Check to return no error, got %v", err)
		}
	})
}

func TestGetConfigPath_WithEnv(t *testing.T) {
	// Reset the global variable
	_configPath = ""

	// Save original env value
	originalPath := os.Getenv("AGENT_FACTORY_CONFIG_PATH")

	// Clean up after test
	defer func() {
		_configPath = ""

		if originalPath != "" {
			os.Setenv("AGENT_FACTORY_CONFIG_PATH", originalPath)
		} else {
			os.Unsetenv("AGENT_FACTORY_CONFIG_PATH")
		}
		// Reinitialize env after cleanup
		cenvhelper.InitEnvForTest()
	}()

	// Set the environment variable
	os.Setenv("AGENT_FACTORY_CONFIG_PATH", "/custom/config/path")

	// Reinitialize cenvhelper to pick up the new env var
	cenvhelper.InitEnvForTest()

	path := GetConfigPath()

	if path != "/custom/config/path" {
		t.Errorf("Expected '/custom/config/path', got '%s'", path)
	}
}

func TestGetConfigPath_Cached(t *testing.T) {
	// Set the global variable directly
	_configPath = "/cached/path"

	// Even if we set an env var, it should return the cached value
	os.Setenv("CONFIG_PATH", "/custom/path")
	defer os.Unsetenv("CONFIG_PATH")

	path := GetConfigPath()

	if path != "/cached/path" {
		t.Errorf("Expected '/cached/path', got '%s'", path)
	}
}

func TestConfig_String_WithNilPtr(t *testing.T) {
	t.Run("config with nil optional fields", func(t *testing.T) {
		config := &Config{
			Project: Project{
				Host: "localhost",
				Port: 8080,
			},
			DB: DBConf{
				UserName: "user",
			},
			// Leave other fields as nil
		}

		str := config.String()
		if str == "" {
			t.Error("Expected String to return a non-empty string")
		}
	})
}

func TestConfig_Check_WithValidLanguage(t *testing.T) {
	t.Run("valid simplified chinese", func(t *testing.T) {
		config := &Config{
			Project: Project{
				Host:     "localhost",
				Port:     8080,
				Language: rest.SimplifiedChinese,
			},
		}

		err := config.Check()
		if err != nil {
			t.Errorf("Expected Check to return no error, got %v", err)
		}
	})
}

func TestConfig_Check_WithInvalidLanguage(t *testing.T) {
	t.Run("invalid language", func(t *testing.T) {
		config := &Config{
			Project: Project{
				Host:     "localhost",
				Port:     8080,
				Language: "",
			},
		}

		err := config.Check()
		if err == nil {
			t.Error("Expected Check to return an error for invalid language")
		}
	})
}

func TestProject_Check(t *testing.T) {
	t.Run("valid simplified chinese", func(t *testing.T) {
		project := Project{
			Host:     "localhost",
			Port:     8080,
			Language: rest.SimplifiedChinese,
		}

		err := project.Check()
		if err != nil {
			t.Errorf("Expected Check to return no error, got %v", err)
		}
	})

	t.Run("valid american english", func(t *testing.T) {
		project := Project{
			Host:     "localhost",
			Port:     8080,
			Language: rest.AmericanEnglish,
		}

		err := project.Check()
		if err != nil {
			t.Errorf("Expected Check to return no error, got %v", err)
		}
	})

	t.Run("empty language", func(t *testing.T) {
		project := Project{
			Host:     "localhost",
			Port:     8080,
			Language: "",
		}

		err := project.Check()
		if err == nil {
			t.Error("Expected Check to return an error for empty language")
		}
	})

	t.Run("invalid language", func(t *testing.T) {
		project := Project{
			Host:     "localhost",
			Port:     8080,
			Language: "invalid",
		}

		err := project.Check()
		if err == nil {
			t.Error("Expected Check to return an error for invalid language")
		}
	})
}

func TestConfig_Check_InvalidProject(t *testing.T) {
	t.Run("invalid project config", func(t *testing.T) {
		config := &Config{
			Project: Project{
				Host:     "localhost",
				Port:     8080,
				Language: "",
			},
		}

		err := config.Check()
		if err == nil {
			t.Error("Expected Check to return an error for invalid project config")
		}
	})
}

func TestConfig_GetDefaultLanguage_Empty(t *testing.T) {
	t.Run("empty language", func(t *testing.T) {
		config := &Config{
			Project: Project{
				Language: "",
			},
		}

		lang := config.GetDefaultLanguage()
		if lang != "" {
			t.Errorf("Expected empty language, got %v", lang)
		}
	})
}

func TestConfig_String_WithFullConfig(t *testing.T) {
	t.Run("full config string representation", func(t *testing.T) {
		config := &Config{
			Project: Project{
				Host:        "localhost",
				Port:        8080,
				Language:    rest.SimplifiedChinese,
				LoggerLevel: 1,
				LogFile:     "/var/log/app.log",
			},
			DB: DBConf{
				UserName: "user",
				Password: "pass",
				DBHost:   "localhost",
				DBPort:   3306,
				DBName:   "testdb",
			},
		}

		str := config.String()
		if str == "" {
			t.Error("Expected String to return a non-empty string")
		}
		// Verify the string contains expected parts
		if !contains(str, "Config") {
			t.Error("Expected String to contain 'Config'")
		}
	})
}

func TestConfig_String_WithAllFields(t *testing.T) {
	t.Run("config with all optional fields", func(t *testing.T) {
		config := &Config{
			Project: Project{
				Host:        "0.0.0.0",
				Port:        30777,
				Language:    rest.AmericanEnglish,
				LoggerLevel: 1,
			},
			DB: DBConf{
				UserName: "testuser",
				DBName:   "testdb",
			},
			Hydra: HydraCfg{
				UserMgnt: UserMgntCfg{
					Host: "user-mgmt.local",
					Port: 8080,
				},
			},
			ModelFactory: &ModelFactoryConf{
				LLM: LLMConf{
					DefaultModelName: "gpt-4",
				},
			},
		}

		str := config.String()
		if str == "" {
			t.Error("Expected String to return a non-empty string")
		}

		if !contains(str, "Config") {
			t.Error("Expected String to contain 'Config'")
		}
	})
}

func TestBaseDefConfig_DBConfig(t *testing.T) {
	t.Run("default db config values", func(t *testing.T) {
		config := BaseDefConfig()

		if config.DB.UserName != "anyshare" {
			t.Errorf("Expected DB.UserName to be 'anyshare', got '%s'", config.DB.UserName)
		}

		if config.DB.DBName != "dip_data_agent" {
			t.Errorf("Expected DB.DBName to be 'dip_data_agent', got '%s'", config.DB.DBName)
		}

		if config.DB.DBPort != 3330 {
			t.Errorf("Expected DB.DBPort to be 3330, got %d", config.DB.DBPort)
		}

		if config.DB.Charset != "utf8mb4" {
			t.Errorf("Expected DB.Charset to be 'utf8mb4', got '%s'", config.DB.Charset)
		}
	})
}

func TestBaseDefConfig_RedisConfig(t *testing.T) {
	t.Run("default redis config values", func(t *testing.T) {
		config := BaseDefConfig()

		if config.Redis.DB != 3 {
			t.Errorf("Expected Redis.DB to be 3, got %d", config.Redis.DB)
		}
	})
}

func TestConfig_MqCfgPath(t *testing.T) {
	t.Run("mq config path is set", func(t *testing.T) {
		config := BaseDefConfig()

		if config.MqCfgPath == "" {
			t.Error("Expected MqCfgPath to be set")
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
