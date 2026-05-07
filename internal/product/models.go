package product

import (
	"strings"
	"unicode"
)

type Product struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Brand    string `json:"brand"`
	Store    string `json:"store"`
	Category string `json:"category"`
}

func capitalize(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))

	if s == "" {
		return ""
	}

	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])

	return string(r)
}

func (p *Product) Normalize() {
	p.Name = capitalize(p.Name)
	p.Brand = capitalize(p.Brand)
	p.Store = capitalize(p.Store)
	p.Category = capitalize(p.Category)
}
