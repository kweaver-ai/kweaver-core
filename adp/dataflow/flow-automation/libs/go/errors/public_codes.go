package errors

// 系统默认错误
const (
	// 公共错误码
	PErrorBadRequest          = "BadRequest"
	PErrorUnauthorized        = "Unauthorized"
	PErrorForbidden           = "Forbidden"
	PErrorNotFound            = "NotFound"
	PErrorMethodNotAllowed    = "MethodNotAllowed"
	PErrorConflict            = "Conflict"
	PErrorInternalServerError = "InternalServerError"
	PErrorNotImplemented      = "NotImplemented"
	PErrorServiceUnavailable  = "ServiceUnavailable"
	PErrorLoopDetected        = "LoopDetected"
)
