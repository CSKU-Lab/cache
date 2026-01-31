package constants

import "errors"

var (
	CONFIG_NOT_FOUND = errors.New("configuration not found")
	NO_CACHE_CONN    = errors.New("no cache connection established")
)
