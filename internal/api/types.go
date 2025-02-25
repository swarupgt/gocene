package api

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
