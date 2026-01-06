package store

import "errors"

var (
	ErrDocumentNotFound error = errors.New("document not found")
	ErrCannotEncodeDoc  error = errors.New("could not encode given document")
	ErrDocFileWrite     error = errors.New("error writing doc bytes to segment file")

	ErrIdxNameExists   error = errors.New("index name already exists")
	ErrIdxDoesNotExist error = errors.New("index with specified name does not exist")

	ErrNotLeader error = errors.New("node not a leader")
)
