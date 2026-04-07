package httpproxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForward_InvalidJSONResponse(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer ts.Close()

	type Req struct {
		Key string `json:"key"`
	}

	type Res struct {
		Val string `json:"val"`
	}

	proxy := NewJSONPostProxy[Req, Res](ts.URL)
	proxy.Client = ts.Client()

	_, err := proxy.Forward(Req{Key: "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JSON解析失败")
}

func TestForward_NonOKStatusCode(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer ts.Close()

	type Req struct {
		Key string `json:"key"`
	}

	type Res struct {
		Val string `json:"val"`
	}

	proxy := NewJSONPostProxy[Req, Res](ts.URL)
	proxy.Client = ts.Client()

	_, err := proxy.Forward(Req{Key: "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}
