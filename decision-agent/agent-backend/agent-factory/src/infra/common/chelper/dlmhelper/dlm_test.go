package dlmhelper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultDlmConf(t *testing.T) {
	t.Parallel()

	t.Run("returns valid config with default values", func(t *testing.T) {
		t.Parallel()

		redisKeyPrefix := "test:dlm:"

		conf := GetDefaultDlmConf(redisKeyPrefix)

		assert.NotNil(t, conf)
		assert.Equal(t, redisKeyPrefix, conf.RedisKeyPrefix)
		assert.NotNil(t, conf.Options)
		assert.NotNil(t, conf.Logger)
		assert.NotNil(t, conf.DeleteValueFunc)
		assert.Greater(t, conf.WatchDogInterval, time.Duration(0))
	})

	t.Run("sets correct expiry and retry values", func(t *testing.T) {
		t.Parallel()

		conf := GetDefaultDlmConf("test:")

		// Verify that the options array is not empty
		assert.NotEmpty(t, conf.Options)

		// WatchDogInterval should be half of expiry (20s / 2 = 10s)
		expectedWatchDogInterval := 10 * time.Second
		assert.Equal(t, expectedWatchDogInterval, conf.WatchDogInterval)
	})

	t.Run("allows custom redis key prefix", func(t *testing.T) {
		t.Parallel()

		customPrefix := "custom:prefix:lock:"

		conf := GetDefaultDlmConf(customPrefix)

		assert.Equal(t, customPrefix, conf.RedisKeyPrefix)
	})
}
