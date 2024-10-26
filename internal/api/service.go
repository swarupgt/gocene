package api

import "gocene/internal/store"

// all api service functions here

type Service struct {
}

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

func (s *Service) GetIndices() (res *GetIndicesResult, err error) {

	res = &GetIndicesResult{}

	for idxName := range store.ActiveIndices {
		res.IndicesList = append(res.IndicesList, idxName)
	}

	return
}
