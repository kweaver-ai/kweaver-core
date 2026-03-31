package rest

const (
	// BadRequest 通用错误码，客户端请求错误
	BadRequest = 400000000
	// InternalServerError 通用错误码，服务端内部错误
	InternalServerError = 500000000
	// Unauthorized 通用错误码，未授权或者授权已过期
	Unauthorized = 401000000
	// URINotExist 通用错误码，请求URI资源不存在
	URINotExist = 404000000
	// Forbidden 通用错误码，禁止访问
	Forbidden = 403000000
	// Conflict 通用错误码，资源冲突
	Conflict = 409000000
)

var (
	commonErrorI18n = map[int]map[string]string{
		BadRequest: {
			Languages[0]: "参数不合法。",
			Languages[1]: "參數不合法。",
			Languages[2]: "Invalid parameter.",
		},
		InternalServerError: {
			Languages[0]: "内部错误",
			Languages[1]: "內部錯誤",
			Languages[2]: "Internal Server Error",
		},
		Unauthorized: {
			Languages[0]: "授权无效",
			Languages[1]: "授權無效",
			Languages[2]: "Unauthorized",
		},
		URINotExist: {
			Languages[0]: "请求uri不存在",
			Languages[1]: "請求uri不存在",
			Languages[2]: "The request uri does not exist",
		},
		Forbidden: {
			Languages[0]: "禁止访问",
			Languages[1]: "禁止訪問",
			Languages[2]: "Forbidden",
		},
		Conflict: {
			Languages[0]: "资源冲突",
			Languages[1]: "資源衝突",
			Languages[2]: "Conflict",
		},
	}
)

func init() {
	Register(commonErrorI18n)
}
