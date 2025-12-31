package store

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// Only parses JSONs of single level and type string at the moment, see how lucene does it for any JSON
func CreateDocumentFromJSON(jsonString string) (doc *Document, err error) {
	// log.Println("inside CreateDocumentFromJSON()")

	// fmt.Println("json doc in CREATEDOCJSON:", jsonString)

	var obj map[string]interface{}

	err = json.Unmarshal([]byte(jsonString), &obj)
	if err != nil {
		// log.Println("ERR IN UNMARSHALLING: ", jsonString)
		return nil, err
	}

	// fmt.Println("obj: ", obj)

	var id int = 0

	doc = NewDocument()

	for key := range obj {
		field := Field{
			ID:              id,
			Name:            key,
			Type:            StringField,
			TokenizerString: " ",
			Value:           fmt.Sprint(obj[key]),
		}

		doc.AddField(field)
		id++
	}

	doc.DocMap = obj

	// fmt.Println("doc after creating:", doc.DocMap["book_name"], doc.DocMap["content"])

	return doc, nil
}

func GetTermsFromPhrase(field, phrase string) (terms []Term) {
	log.Println("inside GetTermsFromPhrase()")
	// Split the search phrase into terms
	strs := strings.Split(phrase, " ")
	for _, str := range strs {
		term := NewTerm(field, str)
		terms = append(terms, term)
	}

	return terms
}

func CreateDocumentFromMap(obj map[string]interface{}) (doc *Document, err error) {
	log.Println("inside CreateDocumentFromMap()")

	var id int = 0

	doc = NewDocument()

	for key := range obj {
		field := Field{
			ID:              id,
			Name:            key,
			Type:            StringField,
			TokenizerString: " ",
			Value:           fmt.Sprint(obj[key]),
		}

		doc.AddField(field)
		id++
	}

	doc.DocMap = obj
	return doc, nil
}
