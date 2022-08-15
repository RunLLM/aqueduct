package operator

import (
	"github.com/dropbox/godropbox/errors"
)

var (
	errWrongNumInputs  = errors.New("Wrong number of operator inputs")
	errWrongNumOutputs = errors.New("Wrong number of operator outputs")
)
