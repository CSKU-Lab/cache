package constants

import "errors"

var (
	CONFIG_NOT_FOUND        = errors.New("configuration not found")
	CACHE_VARIANT_NOT_FOUND = errors.New("cache variant not found")
)
