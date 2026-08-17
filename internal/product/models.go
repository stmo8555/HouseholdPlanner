package product

import (
	"strings"
	"unicode"
)

type Product struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
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
	p.Category = capitalize(p.Category)
}
