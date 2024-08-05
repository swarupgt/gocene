package store

import (
	"errors"
	"fmt"
	"strings"
)

type Index struct {
	// ID             int
	Name            string
	TermDictionary  map[Term]TermData
	Docs            []Document
	Count           int
	CaseSensitivity bool
}

func NewIndex(name string) *Index {
	return &Index{
		Name:            name,
		TermDictionary:  make(map[Term]TermData),
		Docs:            make([]Document, 0),
		Count:           0,
		CaseSensitivity: false,
	}
}

// Needs to use mutex and event queues to handle concurrent events for index.
func (idx *Index) AddDocument(doc *Document) (err error) {

	if doc == nil {
		return errors.New("empty document given")
	}

	doc.ID = idx.Count

	// add doc to list
	idx.Docs = append(idx.Docs, *doc)

	// update term dictionary
	for _, f := range doc.Fields {
		var tokens []string
		if f.TokenizerString != "" {
			tokens = strings.Split(f.Value, f.TokenizerString)
		} else {
			fmt.Println("case")
			if idx.CaseSensitivity {
				fmt.Println("case sense true")

				tokens = append(tokens, f.Value)
			} else {
				fmt.Println("case sense false")
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
				td.DocFrequency[idx.Count] = 1

				idx.TermDictionary[t] = td

			} else {
				// increment the frequency of current term
				if count := idx.TermDictionary[t].DocFrequency[idx.Count]; count == 0 {
					idx.TermDictionary[t].DocFrequency[idx.Count] = 1
				} else {
					idx.TermDictionary[t].DocFrequency[idx.Count] += 1
				}
			}
		}
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

func (idx *Index) DeleteDocument(docID int) (err error) {

	return nil
}
