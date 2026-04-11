package chelper

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAccountTypeFromHeaderMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		headerMap   map[string]string
		wantType    cenum.AccountType
		wantIsExist bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "nil header map",
			headerMap:   nil,
			wantErr:     true,
			errContains: "headerMap is nil",
		},
		{
			name:        "empty header map",
			headerMap:   map[string]string{},
			wantType:    "",
			wantIsExist: false,
			wantErr:     false,
		},
		{
			name: "account type in new header",
			headerMap: map[string]string{
				cenum.HeaderXAccountType.String(): string(cenum.AccountTypeUser),
			},
			wantType:    cenum.AccountTypeUser,
			wantIsExist: true,
			wantErr:     false,
		},
		{
			name: "account type in old header",
			headerMap: map[string]string{
				cenum.HeaderXAccountTypeOld.String(): string(cenum.AccountTypeApp),
			},
			wantType:    cenum.AccountTypeApp,
			wantIsExist: true,
			wantErr:     false,
		},
		{
			name: "new header takes precedence over old",
			headerMap: map[string]string{
				cenum.HeaderXAccountType.String():    string(cenum.AccountTypeUser),
				cenum.HeaderXAccountTypeOld.String(): string(cenum.AccountTypeApp),
			},
			wantType:    cenum.AccountTypeUser,
			wantIsExist: true,
			wantErr:     false,
		},
		{
			name: "invalid account type",
			headerMap: map[string]string{
				cenum.HeaderXAccountType.String(): "invalid_type",
			},
			wantErr:     true,
			errContains: "EnumCheck",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			accountType, isExist, err := GetAccountTypeFromHeaderMap(tt.headerMap)

			if tt.wantErr {
				require.Error(t, err)

				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantType, accountType)
				assert.Equal(t, tt.wantIsExist, isExist)
			}
		})
	}
}

func TestGetAccountTypeFromContext(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		setupContext func(*gin.Context)
		wantType     cenum.AccountType
		wantIsExist  bool
		wantErr      bool
		errContains  string
	}{
		{
			name:         "nil context",
			setupContext: nil,
			wantErr:      true,
			errContains:  "c is nil",
		},
		{
			name: "account type in new header",
			setupContext: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.Header.Set(cenum.HeaderXAccountType.String(), string(cenum.AccountTypeUser))
			},
			wantType:    cenum.AccountTypeUser,
			wantIsExist: true,
			wantErr:     false,
		},
		{
			name: "account type in old header",
			setupContext: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.Header.Set(cenum.HeaderXAccountTypeOld.String(), string(cenum.AccountTypeApp))
			},
			wantType:    cenum.AccountTypeApp,
			wantIsExist: true,
			wantErr:     false,
		},
		{
			name: "new header takes precedence over old",
			setupContext: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.Header.Set(cenum.HeaderXAccountType.String(), string(cenum.AccountTypeUser))
				c.Request.Header.Set(cenum.HeaderXAccountTypeOld.String(), string(cenum.AccountTypeApp))
			},
			wantType:    cenum.AccountTypeUser,
			wantIsExist: true,
			wantErr:     false,
		},
		{
			name: "no account type header",
			setupContext: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
			},
			wantType:    "",
			wantIsExist: false,
			wantErr:     false,
		},
		{
			name: "invalid account type",
			setupContext: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.Header.Set(cenum.HeaderXAccountType.String(), "invalid_type")
			},
			wantErr:     true,
			errContains: "EnumCheck",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var c *gin.Context
			if tt.setupContext != nil {
				c, _ = gin.CreateTestContext(httptest.NewRecorder())
				tt.setupContext(c)
			}

			accountType, isExist, err := GetAccountTypeFromContext(c)

			if tt.wantErr {
				require.Error(t, err)

				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantType, accountType)
				assert.Equal(t, tt.wantIsExist, isExist)
			}
		})
	}
}

func TestGetAccountIDFromHeaderMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		headerMap   map[string]string
		wantID      string
		wantIsExist bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "nil header map",
			headerMap:   nil,
			wantErr:     true,
			errContains: "headerMap is nil",
		},
		{
			name:        "empty header map",
			headerMap:   map[string]string{},
			wantID:      "",
			wantIsExist: false,
			wantErr:     false,
		},
		{
			name: "account ID in new header",
			headerMap: map[string]string{
				cenum.HeaderXAccountID.String(): "account-123",
			},
			wantID:      "account-123",
			wantIsExist: true,
			wantErr:     false,
		},
		{
			name: "account ID in old header",
			headerMap: map[string]string{
				cenum.HeaderXAccountIDOld.String(): "account-456",
			},
			wantID:      "account-456",
			wantIsExist: true,
			wantErr:     false,
		},
		{
			name: "new header takes precedence over old",
			headerMap: map[string]string{
				cenum.HeaderXAccountID.String():    "account-new",
				cenum.HeaderXAccountIDOld.String(): "account-old",
			},
			wantID:      "account-new",
			wantIsExist: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			accountID, isExist, err := GetAccountIDFromHeaderMap(tt.headerMap)

			if tt.wantErr {
				require.Error(t, err)

				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantID, accountID)
				assert.Equal(t, tt.wantIsExist, isExist)
			}
		})
	}
}

func TestGetAccountIDFromContext(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		setupContext func(*gin.Context)
		wantID       string
		wantIsExist  bool
		wantErr      bool
		errContains  string
	}{
		{
			name:         "nil context",
			setupContext: nil,
			wantErr:      true,
			errContains:  "c is nil",
		},
		{
			name: "account ID in new header",
			setupContext: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.Header.Set(cenum.HeaderXAccountID.String(), "account-123")
			},
			wantID:      "account-123",
			wantIsExist: true,
			wantErr:     false,
		},
		{
			name: "account ID in old header",
			setupContext: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.Header.Set(cenum.HeaderXAccountIDOld.String(), "account-456")
			},
			wantID:      "account-456",
			wantIsExist: true,
			wantErr:     false,
		},
		{
			name: "new header takes precedence over old",
			setupContext: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.Header.Set(cenum.HeaderXAccountID.String(), "account-new")
				c.Request.Header.Set(cenum.HeaderXAccountIDOld.String(), "account-old")
			},
			wantID:      "account-new",
			wantIsExist: true,
			wantErr:     false,
		},
		{
			name: "no account ID header",
			setupContext: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
			},
			wantID:      "",
			wantIsExist: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var c *gin.Context
			if tt.setupContext != nil {
				c, _ = gin.CreateTestContext(httptest.NewRecorder())
				tt.setupContext(c)
			}

			accountID, isExist, err := GetAccountIDFromContext(c)

			if tt.wantErr {
				require.Error(t, err)

				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantID, accountID)
				assert.Equal(t, tt.wantIsExist, isExist)
			}
		})
	}
}
