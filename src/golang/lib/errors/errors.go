package errors

import (
	"github.com/dropbox/godropbox/errors"
)

// TODO: Same as Dropbox's IsError(), but ...
func getMessage(err error) string {
	if dbxErr, isDbxErr := err.(errors.DropboxError); isDbxErr {
		return dbxErr.GetMessage()
	} else {
		return err.Error()
	}
}

// TODO: Is() is our custom error comparison method.
func Is(err, target error) bool {
	// If the errors, directly match, we're done.
	if err == target {
		return true
	}

	// Otherwise, we'll have to perform a comparison of the error strings.
	rootErrStr := ""
	rootErr := errors.RootError(err)
	if rootErr != nil {
		rootErrStr = getMessage(rootErr)
	}

	targetMsg := getMessage(target)
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
