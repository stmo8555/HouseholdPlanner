package code

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ23456789"

// Generate returns a random 6-character household code.
func Generate() (string, error) {
	max := big.NewInt(int64(len(charset)))
	var code [6]byte
	for i := range code {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}
	return string(code[:]), nil
}
