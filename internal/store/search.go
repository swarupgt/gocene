package store

import (
	"errors"
	"fmt"
	"log"
	"sort"
)

// Currently only uses the frequency of term to rank the documents.
func (idx *Index) SearchTerm(t Term) (res []RankedDoc, err error) {

	idx.Mutex.RLock()
	defer idx.Mutex.RUnlock()

	// search if index mutex not locked

	td, found := idx.TermDictionary[t]

	if !found {
		return nil, errors.New("no documents contain given term")
	}

	var freqs, docnos []int

	for docno, freq := range td {
		docnos = append(docnos, docno)
		freqs = append(freqs, freq)
	}

	fmt.Println("docnos: ", docnos)

	sort.SliceStable(docnos, func(i, j int) bool {
		return freqs[i] > freqs[j]
	})

	sort.SliceStable(freqs, func(i, j int) bool {
		return freqs[i] > freqs[j]
	})

	for i, docID := range docnos {
		res = append(res, RankedDoc{
			Score: freqs[i],
			DocID: docID,
		})
	}

	return
}

// Future - Use relative positions of terms to rank better.
// Currently uses only total frequency of all words as score to rank results.
func (idx *Index) SearchFullText(terms []Term) (res []RankedDoc, err error) {

	log.Println("inside store SearchFullText()")

	idx.Mutex.RLock()
	defer idx.Mutex.RUnlock()

	var allDocsMap map[int]RankedDoc = make(map[int]RankedDoc)

	for _, term := range terms {
		tempRes, err := idx.SearchTerm(term)
		fmt.Println("found results for term: ", term, " - ", tempRes)
		if err != nil && err.Error() != "no documents contain given term" {
			return nil, err
		}

		for _, iter := range tempRes {
			if _, exists := allDocsMap[iter.DocID]; exists {
				temp := allDocsMap[iter.DocID]
				temp.Score += iter.Score
				allDocsMap[iter.DocID] = temp
			} else {
				allDocsMap[iter.DocID] = RankedDoc{
					Score: iter.Score,
					DocID: iter.DocID,
				}
			}
		}
	}

	var scores []int

	for _, doc := range allDocsMap {
		scores = append(scores, doc.Score)
		res = append(res, doc)
	}

	sort.SliceStable(res, func(i, j int) bool {
		return scores[i] > scores[j]
	})

	return
}
