package validation

import (
	"regexp"
	"strings"

	"github.com/nexdb/nexdb/pkg/errors"
)

// ValidateCollectionName validates the collection name.
func ValidateCollectionName(collection string) error {
	if strings.TrimSpace(collection) == "" {
		return errors.New(errors.ErrCollectionNameIsEmpty)
	}

	collectionNameRegex := regexp.MustCompile(`^[a-z]*$`)
	if !collectionNameRegex.MatchString(collection) {
		return errors.New(errors.ErrCollectionNameIsInvalid)
	}

	return nil
}
