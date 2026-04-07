package anysharedshandler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ==================== Route Registration ====================

func TestAnysharedsHandler_RegPubRouter(t *testing.T) {
	t.Parallel()

	h := &anysharedsHandler{}
	r := gin.New()
	h.RegPubRouter(r.Group("/v3"))
	routes := r.Routes()

	hasGetInfoByPath := false
	hasDirList := false

	for _, route := range routes {
		if route.Method == http.MethodPost && route.Path == "/v3/anyshare7ds/getinfobypath" {
			hasGetInfoByPath = true
		}

		if route.Method == http.MethodPost && route.Path == "/v3/anyshare7ds/dir/list" {
			hasDirList = true
		}
	}

	assert.True(t, hasGetInfoByPath, "should have getinfobypath route")
	assert.True(t, hasDirList, "should have dir/list route")
}

func TestAnysharedsHandler_RegPriRouter(t *testing.T) {
	t.Parallel()

	h := &anysharedsHandler{}
	r := gin.New()
	h.RegPriRouter(r.Group("/internal"))
	// RegPriRouter is empty, should not panic
}

// ==================== RSA Utility Functions ====================

func TestParseRSAPrivateKey_ValidKey(t *testing.T) {
	t.Parallel()

	key, err := parseRSAPrivateKey([]byte(RSA_PRIVATE_KEY))
	assert.NoError(t, err)
	assert.NotNil(t, key)
}

func TestParseRSAPrivateKey_InvalidPEM(t *testing.T) {
	t.Parallel()

	key, err := parseRSAPrivateKey([]byte("not-valid-pem"))
	assert.Error(t, err)
	assert.Nil(t, key)
	assert.Contains(t, err.Error(), "PEM")
}

func TestParseRSAPrivateKey_InvalidKeyBytes(t *testing.T) {
	t.Parallel()

	invalidPEM := `-----BEGIN PRIVATE KEY-----
YWJjZGVmZw==
-----END PRIVATE KEY-----`
	key, err := parseRSAPrivateKey([]byte(invalidPEM))
	assert.Error(t, err)
	assert.Nil(t, key)
}

func TestRSAPrivateDecryptPKCS1_InvalidData(t *testing.T) {
	t.Parallel()

	key, err := parseRSAPrivateKey([]byte(RSA_PRIVATE_KEY))
	assert.NoError(t, err)

	_, err = rsaPrivateDecryptPKCS1(key, []byte("not-encrypted-data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PKCS1")
}

func TestRSAPrivateDecryptOAEP_InvalidData(t *testing.T) {
	t.Parallel()

	key, err := parseRSAPrivateKey([]byte(RSA_PRIVATE_KEY))
	assert.NoError(t, err)

	_, err = rsaPrivateDecryptOAEP(key, []byte("not-encrypted-data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OAEP")
}

// ==================== Handler Error Branches ====================

type asTestLogger struct{}

func (asTestLogger) Infof(string, ...interface{})  {}
func (asTestLogger) Infoln(...interface{})         {}
func (asTestLogger) Debugf(string, ...interface{}) {}
func (asTestLogger) Debugln(...interface{})        {}
func (asTestLogger) Errorf(string, ...interface{}) {}
func (asTestLogger) Errorln(...interface{})        {}
func (asTestLogger) Warnf(string, ...interface{})  {}
func (asTestLogger) Warnln(...interface{})         {}
func (asTestLogger) Panicf(string, ...interface{}) {}
func (asTestLogger) Panicln(...interface{})        {}
func (asTestLogger) Fatalf(string, ...interface{}) {}
func (asTestLogger) Fatalln(...interface{})        {}

func newASTestCtx(body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	return c, recorder
}

func TestGetInfoByPath_InvalidJSON(t *testing.T) {
	t.Parallel()

	h := &anysharedsHandler{logger: asTestLogger{}}

	c, _ := newASTestCtx("not-valid-json")
	h.getInfoByPath(c)
	assert.NotEmpty(t, c.Errors)
}

func TestDirList_InvalidJSON(t *testing.T) {
	t.Parallel()

	h := &anysharedsHandler{logger: asTestLogger{}}

	c, _ := newASTestCtx("not-valid-json")
	h.dirList(c)
	assert.NotEmpty(t, c.Errors)
}

func TestGetInfoByPath_InvalidPassword(t *testing.T) {
	t.Parallel()

	h := &anysharedsHandler{logger: asTestLogger{}}

	body := `{"protocol":"http","host":"localhost","port":8080,"account":"user","password":"not-base64-!!!","namepath":"/test"}`
	c, _ := newASTestCtx(body)
	h.getInfoByPath(c)
	assert.NotEmpty(t, c.Errors)
}

func TestDirList_InvalidPassword(t *testing.T) {
	t.Parallel()

	h := &anysharedsHandler{logger: asTestLogger{}}

	body := `{"protocol":"http","host":"localhost","port":8080,"account":"user","password":"not-base64-!!!","docid":"123"}`
	c, _ := newASTestCtx(body)
	h.dirList(c)
	assert.NotEmpty(t, c.Errors)
}
