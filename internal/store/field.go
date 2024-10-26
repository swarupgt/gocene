package store

type FieldType int

const (
	StringField FieldType = iota
	IntField
)

// Currently only single level. For multilevel fields,
type Field struct {
	ID              int
	Name            string
	Type            FieldType
	TokenizerString string
	Value           string
}
