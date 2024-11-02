package api

import "errors"

var (
	ErrIdxNameExists   error = errors.New("index name already exists")
	ErrIdxDoesNotExist error = errors.New("index with specified name does not exist")
)
