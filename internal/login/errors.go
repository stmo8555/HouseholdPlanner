package login

import "errors"

// ErrNotFound is returned when a user or session lookup yields no row.
var ErrNotFound = errors.New("login: not found")
