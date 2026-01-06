package store

type Document struct {
	ID     int
	Fields []Field
	DocMap map[string]interface{} // for fast retrieval
}

type RankedDoc struct {
	Score int `json:"score"`
	DocID int `json:"doc_id"`
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
