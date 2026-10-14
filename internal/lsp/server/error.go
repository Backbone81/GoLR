package server

import "errors"

var (
	// ErrSessionEndedWithoutShutdown ends a session whose input or exit notification came before the shutdown request.
	ErrSessionEndedWithoutShutdown = errors.New("session ended without shutdown")

	// ErrClientProcessGone ends a session whose client process is gone.
	ErrClientProcessGone = errors.New("client process gone")

	// ErrUnsupportedCharset ends a session on a Content-Type charset other than utf-8.
	ErrUnsupportedCharset = errors.New("unsupported charset")
)

// Error is an error a handler answers its request with, with a code of the specification. Any other error becomes
// InternalError.
type Error struct {
	code    int
	message string
}

// NewError returns an Error with the code and the message.
func NewError(code int, message string) *Error {
	return &Error{
		code:    code,
		message: message,
	}
}

// Error returns the message.
func (e *Error) Error() string {
	return e.message
}
