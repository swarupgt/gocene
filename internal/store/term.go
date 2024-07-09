package store

type Term struct {
	Field string
	Value string
}

type TermData struct {
	Term         Term
	DocFrequency map[int]int // Doc number to frequency mapping
}

func NewTerm(f, v string) Term {
	return Term{
		Field: f,
		Value: v,
	}
}
