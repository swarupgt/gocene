package utils

import (
	"encoding/json"
	"fmt"
	"gocene/internal/store"
)

// Only parses JSONs of single level and type string at the moment, see how lucene does it for any JSON
func CreateDocumentFromJSON(jsonString string) (doc *store.Document, err error) {

	var obj map[string]interface{}

	err = json.Unmarshal([]byte(jsonString), &obj)
	if err != nil {
		return nil, err
	}

	var id int = 0

	doc = store.NewDocument()

	for key := range obj {
		field := store.Field{
			ID:              id,
			Name:            key,
			Type:            store.StringField,
			TokenizerString: " ",
			Value:           fmt.Sprint(obj[key]),
		}

		doc.AddField(field)
		id++
	}

	return doc, nil
}
