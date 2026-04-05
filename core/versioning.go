package core

import (
	"fmt"
	"strings"
)

// VersioningStrategy defines how API versioning is applied.
type VersioningStrategy string

const (
	// URIVersioning prefixes routes with /v{version} (e.g. /v1/users).
	URIVersioning VersioningStrategy = "uri"
	// HeaderVersioning checks a request header for the version (e.g. Accept-Version: 1).
	HeaderVersioning VersioningStrategy = "header"
	// MediaTypeVersioning checks the Accept header for version in media type
	// (e.g. Accept: application/vnd.api.v1+json).
	MediaTypeVersioning VersioningStrategy = "media_type"
)

// VersioningConfig configures the API versioning strategy.
type VersioningConfig struct {
	Strategy VersioningStrategy
	// Header is the header name for HeaderVersioning. Default: "Accept-Version".
	Header string
}

// VersionGuard returns a Guard that checks the request matches the expected version.
// Used internally for header and media type versioning strategies.
func VersionGuard(config VersioningConfig, version string) Guard {
	header := config.Header
	if header == "" {
		header = "Accept-Version"
	}

	return GuardFunc(func(c Context) (bool, error) {
		switch config.Strategy {
		case HeaderVersioning:
			v := c.GetHeader(header)
			if v != version {
				return false, ErrNotFound(fmt.Sprintf("version '%s' not found", v))
			}
			return true, nil

		case MediaTypeVersioning:
			accept := c.GetHeader("Accept")
			versionTag := fmt.Sprintf("v%s", version)
			if !strings.Contains(accept, versionTag) {
				return false, ErrNotFound("unsupported media type version")
			}
			return true, nil

		default:
			// URI versioning doesn't need a guard — handled by prefix
			return true, nil
		}
	})
}
