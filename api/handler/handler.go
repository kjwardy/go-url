package handler

import (
	"regexp"

	"github.com/kjwardy/go-url/api/model"
)

// Handler ...
type Handler struct{}

var (
	urlModel              = &model.URL{}
	urlQueryModel         = &model.URLQuery{}
	validateKeyRegexp     = regexp.MustCompile("^[\\w-]+( +[\\w-]+)*$")
	validateKeyPathRegexp = regexp.MustCompile("^[\\w-]+( +[\\w-]+)*(\\/[\\w-]+( +[\\w-]+)*)*$")
)

// ValidateKey validates a key against the required format
func ValidateKey(key string) bool {
	return validateKeyRegexp.MatchString(key)
}

// ValidateKeyPath validates a key with optional parameters
func ValidateKeyPath(key string) bool {
	return validateKeyPathRegexp.MatchString(key)
}
