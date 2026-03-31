package httphelper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestGClient 创建一个使用 httptest.Server 的 gclient.Client
func newTestClient(handler http.HandlerFunc) (*httptest.Server, *httpClient) {
	server := httptest.NewServer(handler)

	gClient := g.Client()
	gClient.Client = *server.Client()

	c := &httpClient{
		client: gClient,
	}

	return server, c
}

// ==================== Get ====================

func TestHttpClient_Get_Success(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	defer server.Close()

	ctx := context.Background()
	resp, err := c.Get(ctx, server.URL)
	require.NoError(t, err)
	require.NotNil(t, resp)

	defer resp.Close()

	body := resp.ReadAllString()
	assert.Contains(t, body, "ok")
}

func TestHttpClient_GetExpect2xx_Success(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"success"}`))
	})
	defer server.Close()

	ctx := context.Background()
	resp, err := c.GetExpect2xx(ctx, server.URL)
	require.NoError(t, err)
	assert.Contains(t, resp, "success")
}

func TestHttpClient_GetExpect2xx_ServerError(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)

		resp := CommonResp{Code: 500, Message: "server error"}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	ctx := context.Background()
	_, err := c.GetExpect2xx(ctx, server.URL)
	assert.Error(t, err)
}

func TestHttpClient_GetExpect2xxByte_Success(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":"bytes"}`))
	})
	defer server.Close()

	ctx := context.Background()
	resp, err := c.GetExpect2xxByte(ctx, server.URL)
	require.NoError(t, err)
	assert.Contains(t, string(resp), "bytes")
}

func TestHttpClient_GetExpect2xxByte_Non2xx(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`not-json`))
	})
	defer server.Close()

	ctx := context.Background()
	_, err := c.GetExpect2xxByte(ctx, server.URL)
	assert.Error(t, err)
}

// ==================== Post ====================

func TestHttpClient_Post_Success(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"created":true}`))
	})
	defer server.Close()

	ctx := context.Background()
	resp, err := c.Post(ctx, server.URL, `{"name":"test"}`)
	require.NoError(t, err)
	require.NotNil(t, resp)

	defer resp.Close()
}

func TestHttpClient_PostJSONExpect2xx_Success(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"ok"}`))
	})
	defer server.Close()

	ctx := context.Background()
	resp, err := c.PostJSONExpect2xx(ctx, server.URL, `{"name":"test"}`)
	require.NoError(t, err)
	assert.Contains(t, resp, "ok")
}

func TestHttpClient_PostJSONExpect2xxByte_Success(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"ok"}`))
	})
	defer server.Close()

	ctx := context.Background()
	resp, err := c.PostJSONExpect2xxByte(ctx, server.URL, `{"name":"test"}`)
	require.NoError(t, err)
	assert.Contains(t, string(resp), "ok")
}

func TestHttpClient_PostFormExpect2xx_Success(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`form-ok`))
	})
	defer server.Close()

	ctx := context.Background()
	resp, err := c.PostFormExpect2xx(ctx, server.URL, "key=value")
	require.NoError(t, err)
	assert.Contains(t, resp, "form-ok")
}

func TestHttpClient_PostExpect2xx_ServerError(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)

		resp := CommonResp{Code: 500, Message: "error"}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	ctx := context.Background()
	_, err := c.PostExpect2xx(ctx, server.URL, `{}`)
	assert.Error(t, err)
}

// ==================== Put ====================

func TestHttpClient_Put_Success(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"updated":true}`))
	})
	defer server.Close()

	ctx := context.Background()
	resp, err := c.Put(ctx, server.URL, `{"name":"test"}`)
	require.NoError(t, err)
	require.NotNil(t, resp)

	defer resp.Close()
}

func TestHttpClient_PutJSONExpect2xx_Success(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"updated"}`))
	})
	defer server.Close()

	ctx := context.Background()
	resp, err := c.PutJSONExpect2xx(ctx, server.URL, `{"name":"test"}`)
	require.NoError(t, err)
	assert.Contains(t, resp, "updated")
}

func TestHttpClient_PutExpect2xx_ServerError(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)

		resp := CommonResp{Code: 502, Message: "bad gateway"}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	ctx := context.Background()
	_, err := c.PutExpect2xx(ctx, server.URL, `{}`)
	assert.Error(t, err)
}

// ==================== Delete ====================

func TestHttpClient_Delete_Success(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	ctx := context.Background()
	resp, err := c.Delete(ctx, server.URL)
	require.NoError(t, err)
	require.NotNil(t, resp)

	defer resp.Close()
}

func TestHttpClient_DeleteExpect2xx_Success(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`deleted`))
	})
	defer server.Close()

	ctx := context.Background()
	resp, err := c.DeleteExpect2xx(ctx, server.URL)
	require.NoError(t, err)
	assert.Contains(t, resp, "deleted")
}

func TestHttpClient_DeleteExpect2xx_ServerError(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)

		resp := CommonResp{Code: 403, Message: "forbidden"}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	ctx := context.Background()
	_, err := c.DeleteExpect2xx(ctx, server.URL)
	assert.Error(t, err)
}

func TestHttpClient_DeleteExpect2xxWithQueryParams(t *testing.T) {
	t.Parallel()

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "val", r.URL.Query().Get("key"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`ok`))
	})
	defer server.Close()

	ctx := context.Background()
	resp, err := c.DeleteExpect2xxWithQueryParams(ctx, server.URL, map[string]interface{}{"key": "val"})
	require.NoError(t, err)
	assert.Contains(t, resp, "ok")
}

// ==================== errExpect2xx 边界场景 ====================

func TestHttpClient_ErrExpect2xx_LargeBody(t *testing.T) {
	t.Parallel()

	// 返回超过 3000 字节的 non-2xx 响应（触发 body 截断分支）
	largeBody := make([]byte, 4000)
	for i := range largeBody {
		largeBody[i] = 'A'
	}

	server, c := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(largeBody)
	})
	defer server.Close()

	ctx := context.Background()
	_, err := c.GetExpect2xxByte(ctx, server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not 2xx")
}

// ==================== debug 函数 ====================

func TestDebugReqLogger_Log(t *testing.T) {
	t.Parallel()

	l := &debugReqLogger{
		URL:    "http://test.com/api",
		Method: http.MethodPost,
		Data:   map[string]string{"key": "value"},
	}
	// 不 panic 即可
	l.Log()
}

func TestDebugResLogger_Log(t *testing.T) {
	t.Parallel()

	l := &debugResLogger{
		Err:      nil,
		RespBody: []byte(`{"result":"ok"}`),
	}
	// 不 panic 即可
	l.Log()
}

// ==================== errExpect2xxStd ====================

func TestHttpClient_ErrExpect2xxStd_2xx(t *testing.T) {
	t.Parallel()

	c := &httpClient{}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Request:    httptest.NewRequest(http.MethodGet, "/", nil),
	}

	err := c.errExpect2xxStd(resp)
	assert.NoError(t, err)
}

func TestHttpClient_ErrExpect2xxStd_500(t *testing.T) {
	t.Parallel()

	c := &httpClient{}
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Request:    httptest.NewRequest(http.MethodGet, "/test", nil),
	}

	err := c.errExpect2xxStd(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not 2xx")
}
