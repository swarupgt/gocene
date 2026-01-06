package api

import (
	"gocene/internal/store"
)

// All the HTTP types here

type CreateIndexInput struct {
	Name            string `json:"name" binding:"required"`
	CaseSensitivity bool   `json:"case_sensitivity"`
}

type AddDocumentInput struct {
	// IndexName string
	Data map[string]interface{}
}

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
	Results []store.RankedResultDoc `json:"results"`
	Count   int                     `json:"count"`
}

type JoinInput struct {
	NodeID  string `json:"node_id"`
	Address string `json:""`
}

type StatusResult store.StatusResult
