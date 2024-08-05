package store

type FieldType int

const (
	StringField FieldType = iota
	IntField
)

type Field struct {
	ID              int
	Name            string
	Type            FieldType
	TokenizerString string
	Value           string
}
