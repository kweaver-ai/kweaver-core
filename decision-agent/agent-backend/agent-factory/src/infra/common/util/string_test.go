package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeftTrimEllipsisSize(t *testing.T) {
	t.Parallel()

	t.Run("returns original string when shorter than size", func(t *testing.T) {
		t.Parallel()

		result := LeftTrimEllipsisSize("abc", 10)
		assert.Equal(t, "abc", result)
	})

	t.Run("trims string when longer than size", func(t *testing.T) {
		t.Parallel()

		result := LeftTrimEllipsisSize("123456789", 5)
		assert.Equal(t, "12...", result)
	})

	t.Run("handles exact size match", func(t *testing.T) {
		t.Parallel()

		result := LeftTrimEllipsisSize("12345", 5)
		assert.Equal(t, "12345", result)
	})

	t.Run("handles empty string", func(t *testing.T) {
		t.Parallel()

		result := LeftTrimEllipsisSize("", 10)
		assert.Equal(t, "", result)
	})

	t.Run("panics when size is 3 or less", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			LeftTrimEllipsisSize("test", 3)
		})

		assert.Panics(t, func() {
			LeftTrimEllipsisSize("test", 2)
		})

		assert.Panics(t, func() {
			LeftTrimEllipsisSize("test", 0)
		})

		assert.Panics(t, func() {
			LeftTrimEllipsisSize("test", -1)
		})
	})

	t.Run("handles unicode characters", func(t *testing.T) {
		t.Parallel()

		result := LeftTrimEllipsisSize("你好世界欢迎", 5)
		assert.Equal(t, "你好...", result)
	})
}
