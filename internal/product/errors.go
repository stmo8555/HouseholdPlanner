package product

import "errors"

// ErrNotFound is returned when a product lookup yields no row.
var ErrNotFound = errors.New("product: not found")
