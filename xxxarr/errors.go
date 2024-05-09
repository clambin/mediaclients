package xxxarr

import "errors"

var _ error = &ErrInvalidJSON{}

type ErrInvalidJSON struct {
	Err  error
	Body []byte
}

func (e *ErrInvalidJSON) Error() string {
	return "parse: " + e.Err.Error()
}

func (e *ErrInvalidJSON) Is(target error) bool {
	var err *ErrInvalidJSON
	return errors.As(target, &err)
}

func (e *ErrInvalidJSON) Unwrap() error {
	return e.Err
}
