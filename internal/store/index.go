package store

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
)

var ActiveIndices map[string]*Index

type Index struct {
	// ID             int
	Name            string
	TermDictionary  map[Term]TermData
	Docs            []Document
	Count           int // tracks doc ID
	CaseSensitivity bool

	Mutex sync.RWMutex
}

// load persistent indices here
func Init() {
	ActiveIndices = make(map[string]*Index)
}

func NewIndex(name string, cs bool) *Index {
	return &Index{
		Name:            name,
		TermDictionary:  make(map[Term]TermData),
		Docs:            make([]Document, 0),
		Count:           0,
		CaseSensitivity: cs,
	}
}

// Needs to use mutex to handle concurrent events for index.
func (idx *Index) AddDocument(doc *Document) (id int, err error) {

	idx.Mutex.Lock()
	defer idx.Mutex.Unlock()

	if doc == nil {
		return 0, errors.New("empty document given")
	}

	idx.Count++
	doc.ID = idx.Count

	// add doc to list
	idx.Docs = append(idx.Docs, *doc)

	// update term dictionary
	err = idx.UpdateTermDictionary(doc.ID)
	if err != nil {
		return 0, err
	}

	log.Println("term dictionary updated")

	// Save to disk, handle concurrently later
	err = SaveIndexToPersistentMemory(idx)
	return idx.Count, err
}

func (idx *Index) GetAllDocuments() (docs []Document) {
	return idx.Docs
}

// improve lookup time
func (idx *Index) GetDocument(id int) (doc Document, err error) {

	idx.Mutex.RLock()
	defer idx.Mutex.RUnlock()

	for _, docIter := range idx.Docs {
		if docIter.ID == id {
			return docIter, nil
		}
	}

	return Document{}, ErrDocumentNotFound
}

func (idx *Index) GetDocumentCount() int {
	idx.Mutex.RLock()
	defer idx.Mutex.RUnlock()

	return len(idx.Docs)
}

func (idx *Index) GetTermsAndFreqFromDocNo(docNo int) (terms []Term, counts []int) {

	idx.Mutex.RLock()
	defer idx.Mutex.RUnlock()

	for term, termData := range idx.TermDictionary {
		terms = append(terms, term)
		counts = append(counts, termData[docNo]) // return counts of first document
	}

	return
}

// Should this even exist?
func (idx *Index) ModifyDocument(id int, fs []Field) (err error) {
	// do binary search later

	idx.Mutex.Lock()
	defer idx.Mutex.Unlock()

	for i, iter := range idx.Docs {
		if iter.ID == id {
			for _, f := range fs {
				for j, jiter := range idx.Docs[i].Fields {
					if f.Name == jiter.Name {
						//update field value, type and tokeniser string
						idx.Docs[i].Fields[j].Value = f.Value
						idx.Docs[i].Fields[j].Type = f.Type
						idx.Docs[i].Fields[j].TokenizerString = f.TokenizerString
					}
				}
			}
		}
	}

	//update term dictionary
	return idx.UpdateTermDictionary(id)
}

// finish later
func (idx *Index) DeleteDocument(docID int) (err error) {

	idx.Mutex.Lock()
	defer idx.Mutex.Unlock()

	return nil
}

// assumes doc is the updated version
// TODO: send document instead of docID for speed
func (idx *Index) UpdateTermDictionary(docID int) (err error) {

	// get the doc
	var doc Document
	var found bool = false

	for _, docIter := range idx.Docs {
		if docIter.ID == docID {
			doc = docIter
			found = true
			break
		}
	}

	if !found {
		return ErrDocumentNotFound
	}

	fmt.Println("doc retrieved")

	//add the new terms to the term dictionary

	for _, f := range doc.Fields {
		var tokens []string
		if f.TokenizerString != "" {
			tokens = strings.Split(f.Value, f.TokenizerString)
		} else {
			// fmt.Println("case")
			if idx.CaseSensitivity {
				// fmt.Println("case sense true")

				tokens = append(tokens, f.Value)
			} else {
				// fmt.Println("case sense false")
				tokens = append(tokens, strings.ToLower(f.Value))
			}
		}

		for _, token := range tokens {
			t := Term{
				Field: f.Name,
				Value: token,
			}

			// Check if TermData available for the given term
			if _, exists := idx.TermDictionary[t]; !exists {
				// create term data and add to map
				var td TermData = make(map[int]int)
				td[docID] = 1
				idx.TermDictionary[t] = td

			} else {
				// increment the frequency of current term
				idx.TermDictionary[t][docID] += 1
			}
		}
	}

	return nil
}
