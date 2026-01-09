package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Only parses JSONs of single level and type string at the moment, see how lucene does it for any JSON
func CreateDocumentFromJSON(jsonString string) (doc *Document, err error) {

	var obj map[string]interface{}
	err = json.Unmarshal([]byte(jsonString), &obj)
	if err != nil {
		return nil, err
	}

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

// Only parses map of a single level.
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

// Gets usable terms from field and search string input.
func GetTermsFromPhrase(field, phrase string) (terms []Term) {
	log.Println("inside GetTermsFromPhrase()")

	strs := strings.Split(phrase, " ")
	for _, str := range strs {
		term := NewTerm(field, str)
		terms = append(terms, term)
	}

	return terms
}

// Send a HTTP Raft Join request to the leader.
func JoinLeaderAsFollower(httpAddr, joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"node_address": raftAddr, "node_id": nodeID, "http_address": httpAddr})
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
