package store

import (
	"errors"
	"strings"
	"sync"
)

var ActiveIndices map[string]*Index

type Index struct {
	// ID             int
	Name            string
	TermDictionary  map[Term]TermData
	Docs            []Document
	Count           int
	CaseSensitivity bool

	Mutex sync.Mutex
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
func (idx *Index) AddDocument(doc *Document) (err error) {

	if doc == nil {
		return errors.New("empty document given")
	}

	doc.ID = idx.Count

	// add doc to list
	idx.Docs = append(idx.Docs, *doc)

	// update term dictionary
	err = idx.UpdateTermDictionary(doc.ID)
	if err != nil {
		return err
	}

	idx.Count++

	// Save to disk, handle concurrently later
	err = SaveIndexToPersistentMemory(idx)
	return err
}

func (idx *Index) GetAllDocuments() (docs []Document) {
	return idx.Docs
}

// improve lookup time
func (idx *Index) GetDocument(id int) (doc Document, err error) {

	for _, docIter := range idx.Docs {
		if docIter.ID == id {
			return docIter, nil
		}
	}

	return Document{}, errors.New("document not found")
}

func (idx *Index) GetDocumentCount() int {
	return len(idx.Docs)
}

func (idx *Index) GetTermsAndFreqFromDocNo(docNo int) (terms []Term, counts []int) {

	for term, termData := range idx.TermDictionary {
		terms = append(terms, term)
		counts = append(counts, termData.DocFrequency[docNo]) // return counts of first document
	}

	return
}

func (idx *Index) ModifyDocument(id int, fs []Field) (err error) {
	// do binary search later

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

	return nil
}

// assumes doc is the updated version
func (idx *Index) UpdateTermDictionary(docID int) (err error) {

	// remove old terms from the dict using docID
	for t := range idx.TermDictionary {
		delete(idx.TermDictionary[t].DocFrequency, docID)
	}

	doc, err := idx.GetDocument(docID)
	if err != nil {
		return err
	}

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
			if td, exists := idx.TermDictionary[t]; !exists {
				// create term data var and add to map

				td = TermData{
					Term:         t,
					DocFrequency: make(map[int]int),
				}
				td.DocFrequency[docID] = 1

				idx.TermDictionary[t] = td

			} else {
				// increment the frequency of current term
				if count := idx.TermDictionary[t].DocFrequency[docID]; count == 0 {
					idx.TermDictionary[t].DocFrequency[docID] = 1
				} else {
					idx.TermDictionary[t].DocFrequency[docID] += 1
				}
			}
		}
	}

	return nil
}
