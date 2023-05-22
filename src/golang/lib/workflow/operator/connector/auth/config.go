package auth

import (
	"context"

	"github.com/dropbox/godropbox/errors"
)

type configType string

const (
	staticConfigType configType = "staticConfig"
	oauthConfigType  configType = "oauthConfig"
	UserField                   = "user"
	GithubUserField             = "github_user"
)

// parseConfigType parses the specified string into a configType.
// It also returns an error, if any.
func parseConfigType(cType string) (configType, error) {
	c := configType(cType)
	switch c {
	case staticConfigType, oauthConfigType:
		return c, nil
	default:
		return "", errors.Newf("Unknown config type: %v", c)
	}
}

// Config stores the authentication credentials needed by resource connectors.
type Config interface {
	// getType returns the configType of the Config.
	getType() configType

	// Marshal JSON encodes Config into the format that the connector expects.
	Marshal() ([]byte, error)

	// PublicConfig returns a map[string]string of the non-sensitive (i.e. not passwords or tokens)
	// information stored in the Config. This information can safely be stored in plaintext
	// in a database.
	PublicConfig() map[string]string

	// Refresh refreshes the Config if necessary. It returns a bool indicating whether
	// the Config was updated. It also returns an error, if any.
	Refresh(ctx context.Context) (bool, error)
}
