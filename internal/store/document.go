package store

type Document struct {
	Fields []Field
}

func NewDocument() *Document {
	return &Document{
		Fields: make([]Field, 0),
	}
}

func (doc *Document) AddField(f Field) {
	doc.Fields = append(doc.Fields, f)
}

func (doc *Document) Get(s string) string {
	for _, f := range doc.Fields {
		if f.Name == s {
			return f.Value
		}
	}

	return ""
}
