package store

import "strings"

type Term string
type TermData map[int]int // Doc number to frequency mapping

// Term - "field,value"
func NewTerm(f, v string) Term {
	return Term(strings.Join([]string{f, v}, ","))
}
