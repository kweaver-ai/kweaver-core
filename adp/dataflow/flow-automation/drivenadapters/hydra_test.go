package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-playground/assert/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockHydraAdmin(clients *HttpClientMock, client *http.Client) HydraAdmin {
	return &hydraAdmin{
		adminAddress: "http://localhost:8080",
		client:       client,
		httpClient:   clients.httpClient,
	}
}

func TestIntrospect(t *testing.T) {
	Convey("TestIntrospect", t, func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "POST")
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"active":true, "sub":"u1", "client_id":"c1", "exp":200, "iat":100, "ext":{"udid":"device1"}}`)
		}))
		defer server.Close()

		ha := &hydraAdmin{
			adminAddress: server.URL,
			client:       server.Client(),
		}

		info, err := ha.Introspect(context.Background(), "token")
		assert.Equal(t, err, nil)
		assert.Equal(t, info.Active, true)
		assert.Equal(t, info.UserID, "u1")
		assert.Equal(t, info.UdID, "device1")
	})
}

func TestUpdateClientHydra(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	ha := NewMockHydraAdmin(httpClient, nil)
	ctx := context.Background()

	Convey("TestUpdateClient", t, func() {
		httpClient.httpClient.EXPECT().Put(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, nil, nil)
		code, err := ha.UpdateClient(ctx, "id", "name", "secret", "red", "logout")
		assert.Equal(t, err, nil)
		assert.Equal(t, code, 200)
	})
}

func TestRequestTokenWithCredential(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"access_token":"t1", "expires_in":3600}`)
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	hp := &hydraPublic{
		publicAddress:             u.Host,
		contentType:               "Content-Type",
		applicationFormUrlencoded: "application/x-www-form-urlencoded",
		authorization:             "Authorization",
		basic:                     "Basic",
		client:                    server.Client(),
		useHTTPS:                  false,
	}

	Convey("TestRequestTokenWithCredential", t, func() {
		info, code, err := hp.RequestTokenWithCredential("id", "secret", []string{"all"})
		assert.Equal(t, err, nil)
		assert.Equal(t, code, 200)
		assert.Equal(t, info.Token, "t1")
		assert.Equal(t, info.ExpiresIn, 3600)
	})
}

func TestRegisterClient(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	hp := &hydraPublic{
		authPublicAddress: "localhost:8081",
		httpClient:        httpClient.httpClient1,
	}

	Convey("TestRegisterClient", t, func() {
		mockResp := map[string]interface{}{
			"client_id":     "c1",
			"client_secret": "s1",
		}
		httpClient.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
		cid, secret, code, err := hp.RegisterClient("name", "red", "logout")
		assert.Equal(t, err, nil)
		assert.Equal(t, code, 200)
		assert.Equal(t, cid, "c1")
		assert.Equal(t, secret, "s1")
	})
}
