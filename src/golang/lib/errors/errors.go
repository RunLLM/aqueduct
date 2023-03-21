package errors

import (
	"github.com/dropbox/godropbox/errors"
)

// Is is our custom error comparison method. The candidate error can be arbitrarily wrapped,
// but the root error is one we care about and want to compare.
//
// The assumption for the target error is that it is unwrapped, hence why we do no such unwrapping
// for that error. This is because the target error is expected to be simple. Eg. `db.ErrNotFound`.
//
// To compare errors, we compare the root error message with the target error message. Any stack traces
// provided by Dropbox errors are excluded.
func Is(err, target error) bool {
	// If the errors, directly match, we're done.
	if err == target {
		return true
	}

	// Otherwise, we must perform a comparison of the error strings.
	rootErrStr := ""
	rootErr := errors.RootError(err)
	if rootErr != nil {
		rootErrStr = errors.GetMessage(rootErr)
	}

	targetMsg := errors.GetMessage(target)
	return rootErrStr == targetMsg
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
