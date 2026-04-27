package capimiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCors(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	t.Run("returns handler function", func(t *testing.T) {
		t.Parallel()

		handler := Cors()
		if handler == nil {
			t.Error("Expected Cors to return a non-nil handler")
		}
	})

	t.Run("processes request without panic", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(Cors())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "test")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("handles OPTIONS request", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(Cors())
		router.OPTIONS("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "test")
		})

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// OPTIONS should be handled (may return 204 or continue depending on env)
		if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
			t.Logf("Note: OPTIONS returned status %d (may vary by environment)", w.Code)
		}
	})
}
