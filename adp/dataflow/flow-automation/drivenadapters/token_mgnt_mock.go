package drivenadapters

type mockAppTokenMgnt struct{}

func (m *mockAppTokenMgnt) GetAppToken(tokenIn string) (string, int64, error) {
	tokenInfo := mockHydraTokenInfo()
	return tokenInfo.Token, int64(tokenInfo.ExpiresIn), nil
}
