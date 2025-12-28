package config

import (
	"fmt"
	"os"
	"strconv"
)

var (

	// Directories where index files are stored
	IndexDataDirectory    string
	IndexDocListDirectory string

	// service config
	Port               string
	ActiveSegmentCount int
	CaseSensitivity    bool
)

func LoadEnv() {

	IndexDataDirectory = os.Getenv("IDX_DATA_DIR")
	IndexDocListDirectory = os.Getenv("IDX_DOC_LIST_DIR")

	Port = ":" + os.Getenv("PORT")
	ActiveSegmentCount, _ = strconv.Atoi(os.Getenv("MAX_SEGMENT_DOC_COUNT"))
	CaseSensitivity, _ = strconv.ParseBool(os.Getenv("CASE_SENSITIVITY"))

	fmt.Println("test vars", Port, ActiveSegmentCount, IndexDataDirectory)
}
