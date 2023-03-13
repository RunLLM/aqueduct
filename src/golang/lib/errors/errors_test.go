package errors

import (
	stderrors "errors"
	"io"
	"testing"

	"github.com/dropbox/godropbox/errors"
	"github.com/stretchr/testify/require"
)

func TestIs(t *testing.T) {
	type test struct {
		err1     error
		err2     error
		areEqual bool
	}

	tests := []test{
		// Dropbox Errors
		{err1: errors.New("This is a sample error."), err2: errors.New("This is a sample error."), areEqual: true},
		{err1: errors.New("This is a sample error."), err2: errors.New("This is NOT a sample error."), areEqual: false},

		// Standard Errors
		{err1: stderrors.New("This is a sample error."), err2: stderrors.New("This is a sample error."), areEqual: true},
		{err1: stderrors.New("This is a sample error."), err2: stderrors.New("This is NOT a sample error."), areEqual: false},

		// Mix Dropbox and Standard Errors
		{err1: stderrors.New("This is a sample error."), err2: errors.New("This is a sample error."), areEqual: false},

		// Wrapped Errors
		{
			err1:     Wrap(io.EOF, "This is a sample error."),
			err2:     errors.New("This is a sample error."),
			areEqual: true,
		},
	}

	for _, tc := range tests {
		require.Equal(t, tc.areEqual, Is(tc.err1, tc.err2))
	}
}
