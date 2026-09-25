package utils

import "fmt"

// Warning reports an issue which does not stop the build.
//
//nolint:errname // While Warning implements the error interface, it is not a real error.
type Warning struct {
	message string
}

// Warningf returns a warning with the message formatted like fmt.Sprintf.
func Warningf(format string, args ...any) Warning {
	return Warning{message: fmt.Sprintf(format, args...)}
}

// Warning implements error.
var _ error = Warning{}

// Error returns the message of the warning.
func (w Warning) Error() string {
	return w.message
}
