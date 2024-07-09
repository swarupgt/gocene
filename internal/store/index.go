package store

import "strings"

type Index struct {
	// ID             int
	Name           string
	TermDictionary map[Term]TermData
	Docs           []Document
	Count          int
}

func NewIndex() *Index {
	return &Index{
		TermDictionary: make(map[Term]TermData),
		Docs:           make([]Document, 0),
		Count:          0,
	}
}

// Needs to use mutex and event queues to handle concurrent events for index.
func (idx *Index) AddDocument(doc *Document) (err error) {

	// add doc to list
	idx.Docs = append(idx.Docs, *doc)

	// update term dictionary
	for _, f := range doc.Fields {
		var tokens []string
		if f.TokenizerString != "" {
			tokens = strings.Split(f.Value, f.TokenizerString)
		} else {
			tokens = append(tokens, f.Value)
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
	SaveIndexToPersistentMemory(idx)

	return nil
}
