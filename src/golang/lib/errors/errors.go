package errors

import "github.com/dropbox/godropbox/errors"

// AqError exposes additional information about an error
type AqError interface {
	errors.DropboxError
}

// Is reports whether err matches target
func Is(err, target error) bool {
	dboxErr, isErrDbox := err.(errors.DropboxError)
	dboxTarget, isTargetDbox := target.(errors.DropboxError)

	if !isErrDbox && !isTargetDbox {
		// Neither error is a DropboxError
		return errors.IsError(err, target)
	}

	if isErrDbox && isTargetDbox {
		// Both errors are a DropBoxError, so we can only check the outermost
		// error message, without the stack trace.
		return dboxErr.GetMessage() == dboxTarget.GetMessage()
	}

	return false
}

// This returns a new DropboxError initialized with the given message and
// the current stack trace.
func New(msg string) errors.DropboxError {
	return errors.New(msg)
}

// Same as New, but with fmt.Printf-style parameters.
func Newf(format string, args ...interface{}) errors.DropboxError {
	return errors.Newf(format, args...)
}

// Wraps another error in a new DropboxError.
func Wrap(err error, msg string) errors.DropboxError {
	return errors.Wrap(err, msg)
}

// Same as Wrap, but with fmt.Printf-style parameters.
func Wrapf(err error, format string, args ...interface{}) errors.DropboxError {
	return errors.Wrapf(err, format, args...)
}
