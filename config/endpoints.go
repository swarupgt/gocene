package config

const (
	CreateIndexAPI = iota
	GetIndicesAPI
	AddDocumentAPI
	// ModifyDocumentAPI
	// GetDocumentAPI
	// GetAllDocumentsAPI
	// GetIndexDetailsAPI
	// SearchFullTextAPI
	// SearchTermAPI
)

var (
	EndpointsMap map[int]string = map[int]string{
		CreateIndexAPI: "/create_index",
		GetIndicesAPI:  "/indices",
		AddDocumentAPI: "/:idx_name/add_document",
		// ModifyDocumentAPI:  "/:idx_name/modify_document",
		// GetDocumentAPI:     "/:idx_name/get_document",
		// GetAllDocumentsAPI: "/:idx_name/get_all",
		// GetIndexDetailsAPI: "/:idx_name",
		// SearchFullTextAPI:  "/:idx_name/search",
		// SearchTermAPI:      "/:idx_name/search_term",
	}
)
