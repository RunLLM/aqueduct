package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func createErrWithDbxRoot(errMsgs []string) error {
	err := New(errMsgs[0])

	for i := 1; i < len(errMsgs); i++ {
		err = Wrap(err, errMsgs[i])
	}
	return err
}

func createErrWithStdRoot(errMsgs []string) error {
	err := errors.New(errMsgs[0])

	// Assumption: Standard errors are still wrapped with our custom Wrap() function.
	for i := 1; i < len(errMsgs); i++ {
		err = Wrap(err, errMsgs[i])
	}
	return err
}

func TestErrorIs(t *testing.T) {

	// For each test case, we compare all four permutations of DbxError vs Non-DbxError.
	// The results should be the same in all cases.
	type test struct {
		// Sorted so that the first error in the list is the root error and subsequent ones perform wrapping.
		// We always perform the wrapping with our custom Wrap() function.
		err      []string
		target   []string
		areEqual bool
	}

	// NOTE: we never expect the target error to be multi-layered.
	tests := []test{
		// Simple comparison with single-layered error.
		{
			err:      []string{"This is a sample error."},
			target:   []string{"This is a sample error."},
			areEqual: true,
		},
		{
			err:      []string{"This is a sample error."},
			target:   []string{"This is NOT a sample error."},
			areEqual: false,
		},

		// Multiple layered errors.
		{
			err: []string{
				"This is the root error",
				"This is the outermost layer",
			},
			target: []string{
				"This is the root error",
			},
			areEqual: true,
		},
		{
			err: []string{
				"This is the root error",
				"This is the outermost layer",
			},
			target: []string{
				"This is NOT the root error",
			},
			areEqual: false,
		},

		// The target matching the outermost error message is still not equal, since
		// we only compare the root error.
		{
			err: []string{
				"This is the root error",
				"This is the outermost layer",
			},
			target: []string{
				"This is the outermost layer",
			},
			areEqual: false,
		},
	}

	for _, tc := range tests {
		require.Equal(t, tc.areEqual, Is(createErrWithDbxRoot(tc.err), createErrWithDbxRoot(tc.target)))
		require.Equal(t, tc.areEqual, Is(createErrWithDbxRoot(tc.err), createErrWithStdRoot(tc.target)))
		require.Equal(t, tc.areEqual, Is(createErrWithStdRoot(tc.err), createErrWithDbxRoot(tc.target)))
		require.Equal(t, tc.areEqual, Is(createErrWithStdRoot(tc.err), createErrWithStdRoot(tc.target)))
	}
}
