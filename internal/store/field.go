package store

type FieldType int

const (
	String FieldType = iota
	Int
)

type Field struct {
	ID              int
	Name            string
	Type            FieldType
	TokenizerString string
	Value           string
}
