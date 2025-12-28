package config

import "os"

var (

	// Directories where index files are stored
	IndexDataDirectory    = `/home/swarup/workspace/sidep/gocene/app_test_data/data/`
	IndexDocListDirectory = `/home/swarup/workspace/sidep/gocene/app_test_data/doc_lists/`
)

func LoadEnv() {
	IndexDataDirectory = os.Getenv("IDX_DATA_DIR")
	IndexDocListDirectory = os.Getenv("IDX_DOC_LIST_DIR")
}
