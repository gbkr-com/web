package web

import (
	"errors"
)

// Errors for this module.
//
var (
	ErrBadPool  = errors.New("web: bad size for a pool")
	ErrBadQueue = errors.New("web: bad length for a buffered channel")
)
