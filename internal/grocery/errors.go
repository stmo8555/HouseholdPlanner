package grocery

import "errors"

// ErrNotFound is returned when a grocery list or item lookup yields no row.
var ErrNotFound = errors.New("grocery: not found")
