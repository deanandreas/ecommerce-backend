package server

import (
	"encoding/json"
	"errors"
)

var (
	ErrTypeUnmarshal *json.UnmarshalTypeError
	ErrMissingKey    = errors.New("key does not exist")
)
