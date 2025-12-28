package store

import (
	"bufio"
	"io"
	"log"
)

// append doc to a huuge list of all the docs added till now.
// at time of restart, load all of these documents into the index
// not the best way to go about it, but this is just a project to better understand distributed systems.
func (idx *Index) StoreDocumentToDisk(doc *Document) {

}

func (idx *Index) LoadDocumentsIntoIndex() (err error) {
	reader := bufio.NewReader(idx.DocList)

	for {
		docStr, err := reader.ReadString('\n')
		if err == io.EOF {
			if len(docStr) > 0 {
				doc, err := CreateDocumentFromJSON(docStr)
				if err != nil {
					return err
				}
				idx.LoadDocument(doc)
			}
			break
		}

		if err != nil {
			return err
		}

		doc, err := CreateDocumentFromJSON(docStr)
		if err != nil {
			return err
		}
		idx.LoadDocument(doc)

	}

	log.Println("added docs to idx", idx.Name)

	return nil
}
