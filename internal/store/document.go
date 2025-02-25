package store

import "fmt"

type Document struct {
	ID     int
	Fields []Field
	DocMap map[string]interface{} // for fast retrieval
}

type RankedDoc struct {
	Score    int
	Document Document
}

func NewDocument() *Document {
	return &Document{
		Fields: make([]Field, 0),
	}
}

func (doc *Document) AddField(f Field) {

	fmt.Println("field in doc: ", f.Name)

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
