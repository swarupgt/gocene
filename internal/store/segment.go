package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"gocene/config"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
)

type Segment struct {
	Name      string
	TermDict  TermDictionary
	ParentIdx *Index

	// docID to byte offset and length map
	PostingsMap map[int]docPosition

	Docs     *os.File
	Metadata segmentMeta
}

type ActiveSegment struct {
	Seg   *Segment
	Mutex sync.RWMutex

	DocCount int
}

// used for the full-text search
// in memory for now, mmap to disk in the future
type TermDictionary struct {
	dict map[Term]TermData
}

func NewTermDictionary() TermDictionary {
	return TermDictionary{
		dict: make(map[Term]TermData),
	}
}

// doc position in doc file
type docPosition struct {
	byteOffset int
	length     int
	tombstone  bool
}

type segmentMeta struct {
	docCount int
	byteSize int
}

// Returns a new segment with given name, or error if file open unsuccessful
func NewSegment(name string, parentIdx *Index) (*Segment, error) {

	f, err := os.Create(config.SegmentDirectory + name + "_docs.bin")
	if err != nil {
		return nil, err
	}

	return &Segment{
		Name:        name,
		TermDict:    NewTermDictionary(),
		Docs:        f,
		ParentIdx:   parentIdx,
		PostingsMap: make(map[int]docPosition),
		Metadata: segmentMeta{
			docCount: 0,
			byteSize: 0,
		},
	}, nil

}

func NewActiveSegment(name string, parentIdx *Index) (a ActiveSegment, err error) {

	s, err := NewSegment(name, parentIdx)
	if err != nil {
		return
	}

	a.Seg = s

	return
}

// Docs can only be added to an active segment to reduce locking while searching.
func (as *ActiveSegment) AddDocument(doc *Document) (id int, err error) {

	if doc == nil {
		return 0, errors.New("empty document given")
	}

	as.Mutex.Lock()
	defer as.Mutex.Unlock()

	doc.ID = as.Seg.Metadata.docCount + 1

	currByteOffset := as.Seg.Metadata.byteSize

	// add doc to list
	// encode doc to bin and write to end of file

	jsonBytes, err := json.Marshal(doc.DocMap)
	if err != nil {
		return 0, err
	}

	n, err := as.Seg.Docs.Write(jsonBytes)
	if err != nil {
		return 0, ErrDocFileWrite
	}

	as.Seg.PostingsMap[doc.ID] = docPosition{
		byteOffset: int(currByteOffset),
		length:     n,
		tombstone:  false,
	}

	// update term dictionary
	err = as.UpdateTermDictionary(doc)
	if err != nil {
		return 0, err
	}

	log.Println("term dictionary updated")

	as.Seg.Metadata.docCount++
	as.Seg.Metadata.byteSize += n
	return as.Seg.Metadata.docCount, nil
}

// Update active segment's term dictionary
func (as *ActiveSegment) UpdateTermDictionary(doc *Document) (err error) {

	for _, f := range doc.Fields {
		var tokens []string
		if f.TokenizerString != "" {
			tokens = strings.Split(f.Value, f.TokenizerString)
		} else {
			// fmt.Println("case")
			if as.Seg.ParentIdx.CaseSensitivity {
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
			if _, exists := as.Seg.TermDict.dict[t]; !exists {
				// create term data and add to map
				var td TermData = make(map[int]int)
				td[doc.ID] = 1
				as.Seg.TermDict.dict[t] = td

			} else {
				// increment the frequency of current term
				as.Seg.TermDict.dict[t][doc.ID] += 1
			}
		}
	}

	return nil
}

// // Only the active segment is locked since it is mutable.
// func (as *ActiveSegment) SearchFullText(terms []Term) (res []RankedDoc, err error) {
// 	as.Mutex.RLock()
// 	defer as.Mutex.RUnlock()

// 	return as.Seg.SearchFullText(terms)
// }

// Future - Use relative positions of terms to rank better.
// Currently uses only total frequency of all words as score to rank results.
func (seg *Segment) SearchFullText(terms []Term) (res []RankedDoc, err error) {

	log.Println("inside store seg SearchFullText()")

	var allDocsMap map[int]RankedDoc = make(map[int]RankedDoc)

	for _, term := range terms {
		tempRes, err := seg.SearchTerm(term)
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

// Need not lock
func (seg *Segment) SearchTerm(t Term) (res []RankedDoc, err error) {

	td, found := seg.TermDict.dict[t]

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

func (seg *Segment) GetDocument(id int) (docJson string, err error) {

	offset := seg.PostingsMap[id].byteOffset
	length := seg.PostingsMap[id].length

	fmt.Println("offset: ", offset, "length bytes: ", length)

	buff := make([]byte, length)
	_, err = seg.Docs.ReadAt(buff, int64(offset))
	if err != nil {
		return
	}

	docJson = string(buff)
	return
}
