package api

import "gocene/internal/store"

// input structs

type CreateIndexInput struct {
	Name            string `json:"name" binding:"required"`
	CaseSensitivity bool   `json:"case_sensitivity"`
}

type AddDocumentInput struct {
	// IndexName string
	Data map[string]interface{}
}

// result structs
type CreateIndexResult struct {
	Success bool `json:"success"`
}

type GetIndicesResult struct {
	IndicesList []string `json:"indices"`
}

type AddDocumentResult struct {
	DocID   int  `json:"doc_id,omitempty"`
	Success bool `json:"success"`
}

type GetDocumentInput struct {
	DocID int `json:"doc_id" binding:"required"`
}

type GetDocumentResult struct {
	DocID    int                    `json:"doc_id"`
	Document map[string]interface{} `json:"document"`
}

type SearchInput struct {
	SearchField  string `json:"search_field" binding:"required"`
	SearchPhrase string `json:"search_phrase" binding:"required"`
}

type SearchResult struct {
	Results []store.RankedDoc `json:"results"`
	Count   int               `json:"count"`
}
