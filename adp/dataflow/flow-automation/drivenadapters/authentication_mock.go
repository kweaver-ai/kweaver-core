package drivenadapters

type mockAuthentication struct{}

func (m *mockAuthentication) ConfigAuthPerm(appID string) error {
	return nil
}

func (m *mockAuthentication) GetAssertion(userID, token string) (string, error) {
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6Ik1vY2sgVXNlciIsImlhdCI6MTUxNjIzOTAyMn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	return jwt, nil
}
