package errors

import (
	"fmt"

	"github.com/dropbox/godropbox/errors"
	"google.golang.org/grpc/codes"
)

type grpcError struct {
	errors.DropboxError
	Code codes.Code
}

func (e *grpcError) Error() string {
	return fmt.Sprintf("rpc error: code = %s desc = %s", e.Code, e.DropboxError.Error())
}

func NewGRPCError(code codes.Code, errMsg string) error {
	return &grpcError{
		DropboxError: errors.New(errMsg),
		Code:         code,
	}
}

// Wrap a given error with a grpc error code.
// If the error is not a DropboxError, we'll convert it into one.
func WrapInGRPCError(code codes.Code, err error) error {
	if dropboxErr, ok := err.(errors.DropboxError); ok {
		return &grpcError{
			DropboxError: dropboxErr,
			Code:         code,
		}
	}
	return &grpcError{
		DropboxError: errors.New(err.Error()),
		Code:         code,
	}
}
