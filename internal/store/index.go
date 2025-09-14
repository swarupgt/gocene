package store

import (
	"fmt"
	"gocene/config"
	"log"
	"sync"
)

var ActiveIndices map[string]*Index

type Index struct {
	Name     string
	Segments []*Segment

	// separate counter will make merging segments in the future easier
	SegCount int

	CaseSensitivity bool
	Mutex           sync.RWMutex
	As              ActiveSegment
}

// load persistent indices here
func Init() {
	ActiveIndices = make(map[string]*Index)
}

func NewIndex(name string, cs bool) *Index {
	return &Index{
		Name:            name,
		Segments:        nil,
		CaseSensitivity: cs,
	}
}

// Needs to use mutex to handle concurrent events for index.
func (idx *Index) AddDocument(doc *Document) (id int, err error) {

	// if new index, create initial active segment
	if idx.Segments == nil && idx.As.Seg == nil {
		log.Println("creating initial active segment")
		idx.As, err = NewActiveSegment(idx.Name+"_seg_0", idx)
		if err != nil {
			return
		}
	}

	id, err = idx.As.AddDocument(doc)

	// add to segments if active segment full
	if idx.As.Seg.Metadata.docCount >= config.ActiveSegmentCount {
		err = idx.Refresh()
		if err != nil {
			log.Println("could not refresh() index ", idx.Name)
			return
		}
	}

	return
}

// Flushes active segment to segments list to be immutable and search efficient.
func (idx *Index) Refresh() (err error) {

	idx.Mutex.Lock()
	defer idx.Mutex.Unlock()

	idx.Segments = append(idx.Segments, idx.As.Seg)
	idx.SegCount++
	idx.As, err = NewActiveSegment(idx.Name+"seg_"+fmt.Sprint(idx.SegCount), idx)

	return
}

// finish later
func (idx *Index) DeleteDocument(docID int) (err error) {

	return
}
