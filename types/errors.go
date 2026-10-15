// server/types/errors.go
package types

const (
	// 成功
	CodeSuccess = 0

	// 客户端错误 4xx
	CodeParamError       = 400
	CodeUnauthorized     = 401
	CodeForbidden        = 403
	CodeNotFound         = 404
	CodeMethodNotAllowed = 405
	CodeRequestTimeout   = 408
	CodeTooManyRequests  = 429

	// 服务器错误 5xx
	CodeSystemError   = 500
	CodeDatabaseError = 501
	CodeCacheError    = 502
	CodeExternalError = 503

	// 业务错误 10xx
	CodeBusinessError    = 1000
	CodeUserNotFound     = 1001
	CodeUserDisabled     = 1002
	CodePasswordError    = 1003
	CodeTokenExpired     = 1004
	CodeTokenInvalid     = 1005
	CodePermissionDenied = 1006
	CodeResourceNotFound = 1007
	CodeResourceExisted  = 1008
	CodeOperationFailed  = 1009
)

var errorMessages = map[int]string{
	CodeSuccess:          "success",
	CodeParamError:       "Parameter error",
	CodeUnauthorized:     "Unauthorized",
	CodeForbidden:        "Forbidden",
	CodeNotFound:         "Resource not found",
	CodeMethodNotAllowed: "Method not allowed",
	CodeRequestTimeout:   "Request timeout",
	CodeTooManyRequests:  "Too many requests",
	CodeSystemError:      "System error",
	CodeDatabaseError:    "Database error",
	CodeCacheError:       "Cache error",
	CodeExternalError:    "External service error",
	CodeBusinessError:    "Business error",
	CodeUserNotFound:     "User not found",
	CodeUserDisabled:     "User disabled",
	CodePasswordError:    "Invalid password",
	CodeTokenExpired:     "Token expired",
	CodeTokenInvalid:     "Invalid token",
	CodePermissionDenied: "Permission denied",
	CodeResourceNotFound: "Resource not found",
	CodeResourceExisted:  "Resource already exists",
	CodeOperationFailed:  "Operation failed",
}

func GetErrorMessage(code int) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "Unknown error"
}

type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e ApiError) Error() string {
	return e.Message
}
