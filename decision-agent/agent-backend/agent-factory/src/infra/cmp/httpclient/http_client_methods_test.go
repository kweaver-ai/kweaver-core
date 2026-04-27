package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHTTPClient(transport http.RoundTripper) *httpClient {
	return &httpClient{
		client: &http.Client{
			Transport: transport,
			Timeout:   5 * time.Second,
		},
	}
}

func TestHTTPClient_Get_Happy(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestHTTPClient(nil)
	c.client = srv.Client()

	code, body, err := c.Get(srv.URL+"/test", map[string]string{"X-Custom": "val"})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Contains(t, string(body), "ok")
}

func TestHTTPClient_Get_NilClient(t *testing.T) {
	t.Parallel()

	c := &httpClient{client: nil}
	code, body, err := c.Get("http://localhost", nil)
	assert.Error(t, err)
	assert.Equal(t, 0, code)
	assert.Nil(t, body)
}

func TestHTTPClient_Post_Happy(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"1"}`))
	}))
	defer srv.Close()

	c := &httpClient{client: srv.Client()}
	code, body, err := c.Post(srv.URL, nil, map[string]string{"key": "val"})
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, code)
	assert.Contains(t, string(body), "id")
}

func TestHTTPClient_Post_ByteSliceBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	}))
	defer srv.Close()

	c := &httpClient{client: srv.Client()}
	code, body, err := c.Post(srv.URL, nil, []byte(`{"ping":true}`))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "pong", string(body))
}

func TestHTTPClient_Post_NilClient(t *testing.T) {
	t.Parallel()

	c := &httpClient{client: nil}
	code, body, err := c.Post("http://localhost", nil, map[string]string{})
	assert.Error(t, err)
	assert.Equal(t, 0, code)
	assert.Nil(t, body)
}

func TestHTTPClient_Put_Happy(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"updated":true}`))
	}))
	defer srv.Close()

	c := &httpClient{client: srv.Client()}
	code, body, err := c.Put(srv.URL, nil, map[string]string{"x": "y"})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Contains(t, string(body), "updated")
}

func TestHTTPClient_Delete_Happy(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := &httpClient{client: srv.Client()}
	body, err := c.Delete(srv.URL, map[string]string{"Auth": "token"})
	require.NoError(t, err)
	assert.Empty(t, body)
}

func TestHTTPClient_Delete_NilClient(t *testing.T) {
	t.Parallel()

	c := &httpClient{client: nil}
	_, err := c.Delete("http://localhost", nil)
	assert.Error(t, err)
}

func TestHTTPClient_HttpForward_Happy(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"forwarded":true}`))
	}))
	defer srv.Close()

	c := &httpClient{client: srv.Client()}

	origReq, err := http.NewRequest(http.MethodGet, srv.URL, strings.NewReader(""))
	require.NoError(t, err)

	code, body, _, err := c.HttpForward(origReq, srv.URL, http.MethodGet)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Contains(t, string(body), "forwarded")
}

func TestHTTPClient_addHeaders_Empty(t *testing.T) {
	t.Parallel()

	c := &httpClient{}
	req, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
	c.addHeaders(req, nil)
	assert.Empty(t, req.Header)

	c.addHeaders(req, map[string]string{})
	assert.Empty(t, req.Header)
}

func TestHTTPClient_addHeaders_SkipsEmptyValues(t *testing.T) {
	t.Parallel()

	c := &httpClient{}
	req, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
	c.addHeaders(req, map[string]string{
		"X-Valid":   "value",
		"X-Empty":   "",
		"X-Another": "v2",
	})
	assert.Equal(t, "value", req.Header.Get("X-Valid"))
	assert.Equal(t, "", req.Header.Get("X-Empty"))
	assert.Equal(t, "v2", req.Header.Get("X-Another"))
}

func TestHTTPClient_StreamPost_NonOKStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	c := &httpClient{client: srv.Client()}
	ctx := context.Background()

	_, errs, err := c.StreamPost(ctx, srv.URL, nil, []byte(`{}`))
	require.NoError(t, err)

	var gotErr error

	for e := range errs {
		if e != nil {
			gotErr = e
		}
	}

	assert.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "500")
}

func TestHTTPClient_Get_WithHeaders(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := &httpClient{client: srv.Client()}
	code, _, err := c.Get(srv.URL, map[string]string{"Authorization": "Bearer token"})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
}
