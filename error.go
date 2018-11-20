package httpx

import "fmt"

type statusError struct {
	message string
	status  int
}

func (e *statusError) Error() string {
	return e.message
}

func (e *statusError) Status() int {
	return e.status
}

// StatusError is an error type that contains an http status.
type StatusError interface {
	error
	Status() int
}

// Error returns a new error object that includes an http status code.
func Error(status int, message string) error {
	return &statusError{message, status}
}

// Error returns a new error object that includes an http status code
// and a formatted error message.
func Errorf(status int, format string, v ...interface{}) error {
	return Error(status, fmt.Sprintf(format, v...))
}
