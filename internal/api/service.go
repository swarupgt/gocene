package api

import (
	"gocene/internal/store"
	"gocene/internal/utils"
	"log"
)

// all api service functions here

type Service struct {
}

// Creates a new index.
func (s *Service) CreateIndex(inp CreateIndexInput) (res *CreateIndexResult, err error) {

	// return err if same name exists
	if _, ok := store.ActiveIndices[inp.Name]; ok {
		return nil, ErrIdxNameExists
	}

	var idx *store.Index
	if inp.CaseSensitivity {
		idx = store.NewIndex(inp.Name, true)
	} else {
		idx = store.NewIndex(inp.Name, false)
	}

	//add to active index after creating
	store.ActiveIndices[inp.Name] = idx

	res = &CreateIndexResult{
		Success: true,
	}

	return
}

// Gets the list of all active indices on the service.
func (s *Service) GetIndices() (res *GetIndicesResult, err error) {

	res = &GetIndicesResult{}

	for idxName := range store.ActiveIndices {
		res.IndicesList = append(res.IndicesList, idxName)
	}

	return
}

func (s *Service) AddDocument(idxName string, inp AddDocumentInput) (res *AddDocumentResult, err error) {

	log.Println("inside service AddDocument()")

	var idx *store.Index
	var ok bool

	if idx, ok = store.ActiveIndices[idxName]; !ok {
		log.Println(ErrIdxDoesNotExist.Error())
		//index does not exist
		return nil, ErrIdxDoesNotExist
	}

	// create doc and add

	doc, err := utils.CreateDocumentFromMap(inp.Data)
	if err != nil {
		return &AddDocumentResult{
			Success: false,
		}, err
	}

	docId, err := idx.AddDocument(doc)
	if err != nil {
		return &AddDocumentResult{
			Success: false,
		}, err
	}

	res = &AddDocumentResult{
		DocID:   docId,
		Success: true,
	}

	return

}

func (s *Service) GetDocument(idxName string, inp GetDocumentInput) (res *GetDocumentResult, err error) {

	log.Println("inside service GetDocument()")

	var idx *store.Index
	var ok bool

	if idx, ok = store.ActiveIndices[idxName]; !ok {
		log.Println(ErrIdxDoesNotExist.Error())
		//index does not exist
		return nil, ErrIdxDoesNotExist
	}

	doc, err := idx.GetDocument(inp.DocID)
	if err != nil {
		return
	}

	res = &GetDocumentResult{
		DocID:    doc.ID,
		Document: doc.DocMap,
	}

	return
}

func (s *Service) SearchFullText(idxName string, inp SearchInput) (res *SearchResult, err error) {

	log.Println("inside service SearchFullText()")

	var idx *store.Index
	var ok bool

	if idx, ok = store.ActiveIndices[idxName]; !ok {
		log.Println(ErrIdxDoesNotExist.Error())
		//index does not exist
		return nil, ErrIdxDoesNotExist
	}

	terms := utils.GetTermsFromPhrase(inp.SearchField, inp.SearchPhrase)

	rankedDocs, err := idx.SearchFullText(terms)
	if err != nil {
		return nil, err
	}

	res = &SearchResult{
		Results: rankedDocs,
		Count:   len(rankedDocs),
	}

	return
}
