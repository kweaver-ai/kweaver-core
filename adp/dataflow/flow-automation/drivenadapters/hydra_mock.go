package drivenadapters

import (
	"context"
	"net/http"
)

const (
	mockHydraUserID       = "mock-user-id"
	mockHydraClientID     = "mock-client-id"
	mockHydraClientSecret = "mock-client-secret"
	mockHydraAccessToken  = "mock-access-token"
	mockHydraRefreshToken = "mock-refresh-token"
)

type mockHydraAdmin struct{}

func (m *mockHydraAdmin) Introspect(ctx context.Context, token string) (TokenIntrospectInfo, error) {
	return TokenIntrospectInfo{
		Active:      true,
		UserID:      mockHydraUserID,
		UdID:        "mock-device-id",
		LoginIP:     "127.0.0.1",
		VisitorType: "realname",
		ClientType:  "web",
		ClientID:    mockHydraClientID,
		ExpiresIn:   3600,
	}, nil
}

func (m *mockHydraAdmin) UpdateClient(ctx context.Context, id, name, secret, redirectURI, logoutRedirectURI string) (int, error) {
	return http.StatusOK, nil
}

type mockHydraPublic struct{}

func (m *mockHydraPublic) RequestTokenWithCredential(id, secret string, scope []string) (TokenInfo, int, error) {
	return mockHydraTokenInfo(), http.StatusOK, nil
}

func (m *mockHydraPublic) RequestTokenWithRefreshToken(id, secret, token string) (TokenInfo, int, error) {
	return mockHydraTokenInfo(), http.StatusOK, nil
}

func (m *mockHydraPublic) RequestTokenWithCode(id, secret, code, redirect string) (TokenInfo, int, error) {
	return mockHydraTokenInfo(), http.StatusOK, nil
}

func (m *mockHydraPublic) RequestTokenWithAsserts(id, secret, assertion string) (TokenInfo, int, error) {
	return mockHydraTokenInfo(), http.StatusOK, nil
}

func (m *mockHydraPublic) RegisterClient(name, redirectURI, logoutRedirectURI string) (string, string, int, error) {
	return mockHydraClientID, mockHydraClientSecret, http.StatusOK, nil
}

func mockHydraTokenInfo() TokenInfo {
	return TokenInfo{
		Token:        mockHydraAccessToken,
		ExpiresIn:    3600,
		RefreshToken: mockHydraRefreshToken,
	}
}
