package errors

import "fmt"

// VaultError is a custom error type for vault operations
type VaultError struct {
	Code    string
	Message string
	Err     error
}

func (e *VaultError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Error codes
const (
	ErrNotFound       = "NOT_FOUND"
	ErrUnauthorized   = "UNAUTHORIZED"
	ErrInvalidInput   = "INVALID_INPUT"
	ErrDatabaseError  = "DATABASE_ERROR"
	ErrEncryptionFail = "ENCRYPTION_FAILED"
)

func NewVaultError(code, message string) *VaultError {
	return &VaultError{Code: code, Message: message}
}

func NewVaultErrorWithErr(code, message string, err error) *VaultError {
	return &VaultError{Code: code, Message: message, Err: err}
}
