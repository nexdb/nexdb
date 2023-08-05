package auth

import (
	"github.com/nexdb/nexdb/pkg/database"
	"github.com/nexdb/nexdb/pkg/database/cache"
	"github.com/nexdb/nexdb/pkg/errors"
)

// AuthService is a service that handles requests from middleware
// to authenticate users.
type AuthService struct {
	database *database.Database
}

// Authenticate authenticates a user.
func (a *AuthService) Authenticate(key string) error {
	results := a.database.Filter("_api_keys", cache.Query{
		And: []cache.Element{
			{
				Condition: &cache.Condition{
					Field:    "key",
					Operator: cache.Equals,
					Value:    key,
				},
			},
		},
	})
	if len(results) == 0 {
		return errors.New(errors.ErrUnauthorized)
	}

	return nil
}

func New(d *database.Database) *AuthService {
	return &AuthService{
		database: d,
	}
}
