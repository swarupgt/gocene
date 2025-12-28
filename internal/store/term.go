package store

type Term struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type TermData map[int]int // Doc number to frequency mapping

func NewTerm(f, v string) Term {
	return Term{
		Field: f,
		Value: v,
	}
}
