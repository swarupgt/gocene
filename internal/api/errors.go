package api

import "errors"

var (
	ErrIdxNameExists error = errors.New("index name already exists")
)
