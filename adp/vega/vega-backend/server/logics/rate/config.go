// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package rate

import (
	"fmt"
)

// ConcurrencyConfig contains configuration for concurrency control.
type ConcurrencyConfig struct {
	// Enabled controls whether concurrency limiting is active
	Enabled bool `yaml:"enabled"`

	// Global configuration: protects VEGA service itself
	Global GlobalConcurrencyConfig `yaml:"global"`
}

// GlobalConcurrencyConfig configures global concurrency limits.
type GlobalConcurrencyConfig struct {
	// MaxConcurrentQueries is the maximum number of concurrent queries system-wide
	MaxConcurrentQueries int `yaml:"max_concurrent_queries"`
}

// DefaultConcurrencyConfig returns a default concurrency configuration.
func DefaultConcurrencyConfig() ConcurrencyConfig {
	return ConcurrencyConfig{
		Enabled: true,
		Global: GlobalConcurrencyConfig{
			MaxConcurrentQueries: 100,
		},
	}
}

// MergeWithDefaults merges user configuration with defaults.
func (cfg *ConcurrencyConfig) MergeWithDefaults() {
	if !cfg.Enabled {
		return
	}

	// Merge global config
	if cfg.Global.MaxConcurrentQueries == 0 {
		cfg.Global.MaxConcurrentQueries = DefaultConcurrencyConfig().Global.MaxConcurrentQueries
	}
}

// Validate validates the configuration.
func (cfg *ConcurrencyConfig) Validate() error {
	if !cfg.Enabled {
		return nil
	}

	// Validate global config
	if cfg.Global.MaxConcurrentQueries <= 0 {
		return fmt.Errorf("global max_concurrent_queries must be positive")
	}

	return nil
}
