package store

import (
	"fmt"
	"gocene/config"
	"log"
	"os"
	"sync"

	"github.com/minio/minio-go/v7"
)

type Index struct {
	Name     string
	Segments []*Segment

	DocList *os.File
	// separate counter will make merging segments in the future easier
	SegCount  int
	NextDocID int

	CaseSensitivity bool
	Mutex           sync.RWMutex
	As              ActiveSegment

	mc *minio.Client
}

func NewIndex(name string, cs bool) *Index {
	temp := &Index{
		Name:            name,
		Segments:        nil,
		CaseSensitivity: cs,
	}
	var err error
	temp.DocList, err = os.OpenFile(config.IndexDocListDirectory+name, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)

	if err != nil {
		log.Fatalln("could not open file for idx doc list, err: ", err.Error())
	}

	err = os.MkdirAll(config.IndexDataDirectory+"/"+temp.Name, os.ModePerm)
	if err != nil {
		log.Panicln("could not create index, err: ", err.Error())
	}

	return temp
}

// Needs to use mutex to handle concurrent events for index.
func (idx *Index) AddDocument(doc *Document) (id int, err error) {

	idx.NextDocID++

	// if new index, create initial active segment
	if idx.Segments == nil && idx.As.Seg == nil {
		log.Println("creating initial active segment")
		idx.As, err = NewActiveSegment("seg_0", idx)
		if err != nil {
			return
		}
	}

	id, err = idx.As.AddDocument(doc)

	// add to segments if active segment full
	if idx.As.Seg.DocCount >= config.ActiveSegmentCount {
		err = idx.Refresh()
		if err != nil {
			log.Println("could not refresh() index ", idx.Name)
			return
		}
	}

	return
}

// Loading existing documents, without appending
func (idx *Index) LoadDocument(doc *Document) (id int, err error) {

	// if new index, create initial active segment
	if idx.Segments == nil && idx.As.Seg == nil {
		log.Println("creating initial active segment")
		idx.As, err = NewActiveSegment("seg_0", idx)
		if err != nil {
			return
		}
	}

	id, err = idx.As.AddDocument(doc)

	// add to segments if active segment full
	if idx.As.Seg.DocCount >= config.ActiveSegmentCount {
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
	idx.As, err = NewActiveSegment("seg_"+fmt.Sprint(idx.SegCount), idx)

	return
}

// finish later
func (idx *Index) DeleteDocument(docID int) (err error) {

	return
}
