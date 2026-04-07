package httpproxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONPostProxy_Forward_Happy(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req TestRequest

		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "hello", req.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(TestResponse{Result: "ok", Status: 200})
	}))
	defer srv.Close()

	proxy := NewJSONPostProxy[TestRequest, TestResponse](srv.URL)
	proxy.SetClient(srv.Client())

	resp, err := proxy.Forward(TestRequest{Name: "hello", Value: 42})
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Result)
}

func TestJSONPostProxy_Forward_WithToken(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer mytoken", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(TestResponse{Result: "authed"})
	}))
	defer srv.Close()

	proxy := NewJSONPostProxy[TestRequest, TestResponse](srv.URL)
	proxy.SetClient(srv.Client())
	proxy.SetToken("mytoken")

	resp, err := proxy.Forward(TestRequest{Name: "x"})
	require.NoError(t, err)
	assert.Equal(t, "authed", resp.Result)
}

func TestJSONPostProxy_Forward_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("bad gateway"))
	}))
	defer srv.Close()

	proxy := NewJSONPostProxy[TestRequest, TestResponse](srv.URL)
	proxy.SetClient(srv.Client())

	_, err := proxy.Forward(TestRequest{Name: "x"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "502")
}

func TestJSONPostProxy_Forward_InvalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	proxy := NewJSONPostProxy[TestRequest, TestResponse](srv.URL)
	proxy.SetClient(srv.Client())

	_, err := proxy.Forward(TestRequest{Name: "x"})
	assert.Error(t, err)
}

func TestJSONPostProxy_Forward_RequestFailed(t *testing.T) {
	t.Parallel()

	proxy := NewJSONPostProxy[TestRequest, TestResponse]("http://127.0.0.1:1")
	_, err := proxy.Forward(TestRequest{Name: "x"})
	assert.Error(t, err)
}
