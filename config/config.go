package config

import (
	"os"
	"strconv"
	"time"
)

var (

	// Directories where index files are stored
	IndexDataDirectory    string
	IndexDocListDirectory string

	// service config
	Port               string
	ActiveSegmentCount int
	CaseSensitivity    bool

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string

	// raft config
	RaftBootstrap   bool
	RaftJoinAddress string

	MinioDocPathPrefix string = "/docs/"

	RaftTimeout time.Duration = 10 * time.Second
)

func LoadEnv() {

	IndexDataDirectory = os.Getenv("IDX_DATA_DIR")
	IndexDocListDirectory = os.Getenv("IDX_DOC_LIST_DIR")

	Port = ":" + os.Getenv("PORT")
	ActiveSegmentCount, _ = strconv.Atoi(os.Getenv("MAX_SEGMENT_DOC_COUNT"))
	CaseSensitivity, _ = strconv.ParseBool(os.Getenv("CASE_SENSITIVITY"))

	MinioEndpoint = os.Getenv("MINIO_ENDPOINT")
	MinioAccessKey = os.Getenv("MINIO_ACCESS_KEY")
	MinioSecretKey = os.Getenv("MINIO_SECRET_KEY")
	MinioBucket = os.Getenv("MINIO_BUCKET")

	RaftBootstrap, _ = strconv.ParseBool(os.Getenv("GOCENE_BOOTSTRAP"))
	RaftJoinAddress = os.Getenv("GOCENE_JOIN_ADDRESS")
}
