package cconf

import (
	"testing"
)

func TestDBConf_Fields(t *testing.T) {
	t.Run("db config fields are accessible", func(t *testing.T) {
		dbconf := DBConf{
			UserName:         "testuser",
			Password:         "testpass",
			DBHost:           "localhost",
			DBPort:           3306,
			DBName:           "testdb",
			Charset:          "utf8mb4",
			Timeout:          10,
			TimeoutRead:      10,
			TimeoutWrite:     10,
			MaxOpenConns:     30,
			MaxOpenReadConns: 30,
		}

		if dbconf.UserName != "testuser" {
			t.Errorf("Expected UserName to be 'testuser', got '%s'", dbconf.UserName)
		}

		if dbconf.DBPort != 3306 {
			t.Errorf("Expected DBPort to be 3306, got %d", dbconf.DBPort)
		}

		if dbconf.DBName != "testdb" {
			t.Errorf("Expected DBName to be 'testdb', got '%s'", dbconf.DBName)
		}
	})
}

func TestDBConf_DefaultValues(t *testing.T) {
	t.Run("db config default values", func(t *testing.T) {
		dbconf := DBConf{}

		if dbconf.DBName != "" {
			t.Errorf("Expected empty DBName, got '%s'", dbconf.DBName)
		}

		if dbconf.DBPort != 0 {
			t.Errorf("Expected DBPort to be 0, got %d", dbconf.DBPort)
		}
	})
}

func TestDBConf_TimeoutSettings(t *testing.T) {
	t.Run("db config timeout settings", func(t *testing.T) {
		dbconf := DBConf{
			Timeout:      30,
			TimeoutRead:  60,
			TimeoutWrite: 60,
		}

		if dbconf.Timeout != 30 {
			t.Errorf("Expected Timeout to be 30, got %d", dbconf.Timeout)
		}

		if dbconf.TimeoutRead != 60 {
			t.Errorf("Expected TimeoutRead to be 60, got %d", dbconf.TimeoutRead)
		}

		if dbconf.TimeoutWrite != 60 {
			t.Errorf("Expected TimeoutWrite to be 60, got %d", dbconf.TimeoutWrite)
		}
	})
}
