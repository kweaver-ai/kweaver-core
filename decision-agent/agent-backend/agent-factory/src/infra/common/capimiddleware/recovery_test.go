package capimiddleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRecovery(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	t.Run("returns handler function", func(t *testing.T) {
		t.Parallel()

		handler := Recovery()
		if handler == nil {
			t.Error("Expected Recovery to return a non-nil handler")
		}
	})

	t.Run("recovers from panic", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(Recovery())
		router.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		req := httptest.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 500 internal server error after panic recovery
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500 after panic, got %d", w.Code)
		}
	})

	t.Run("handles normal requests without panic", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(Recovery())
		router.GET("/normal", func(c *gin.Context) {
			c.String(http.StatusOK, "normal response")
		})

		req := httptest.NewRequest("GET", "/normal", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for normal request, got %d", w.Code)
		}
	})

	t.Run("recovers from error panic", func(t *testing.T) {
		t.Parallel()

		router := gin.New()
		router.Use(Recovery())
		router.GET("/error-panic", func(c *gin.Context) {
			panic(errors.New("test error"))
		})

		req := httptest.NewRequest("GET", "/error-panic", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 500 after error panic
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500 after error panic, got %d", w.Code)
		}
	})
}
