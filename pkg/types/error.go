package types

import "fmt"

type XiaoYiError struct {
	Code    string
	Message string
	Err     error
}

func (e *XiaoYiError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *XiaoYiError) Unwrap() error {
	return e.Err
}

var (
	ErrNotConnected    = &XiaoYiError{Code: "NOT_CONNECTED", Message: "not connected to server"}
	ErrSessionNotFound = &XiaoYiError{Code: "SESSION_NOT_FOUND", Message: "session not found"}
	ErrConfigInvalid   = &XiaoYiError{Code: "CONFIG_INVALID", Message: "invalid configuration"}
	ErrServerNotReady  = &XiaoYiError{Code: "SERVER_NOT_READY", Message: "server not ready"}
	ErrSendFailed      = &XiaoYiError{Code: "SEND_FAILED", Message: "failed to send message"}
	ErrConnectFailed   = &XiaoYiError{Code: "CONNECT_FAILED", Message: "failed to connect"}
)
