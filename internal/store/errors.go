package store

import "errors"

var (
	ErrDocumentNotFound error = errors.New("document not found")
	ErrCannotEncodeDoc  error = errors.New("could not encode given document")
	ErrDocFileWrite     error = errors.New("error writing doc bytes to segment file")
)
