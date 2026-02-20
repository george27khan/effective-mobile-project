package errors_repo

import (
	"errors"
)

var (
	ErrSubsNotFound = errors.New("subscription not found")
)
