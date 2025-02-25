package store

import "errors"

var (
	ErrDocumentNotFound error = errors.New("document not found")
)
