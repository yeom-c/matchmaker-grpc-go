package helper

import (
	"github.com/pkg/errors"
)

func ErrorWithStack(err string) error {
	return errors.WithStack(errors.New(err))
}
