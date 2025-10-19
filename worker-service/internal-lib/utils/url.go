package utils

import (
	"fmt"
	"net/url"
)

// MustParseURL parses a URL string or panics.
func MustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(fmt.Sprintf("failed to parse URL %s: %v", rawURL, err))
	}
	return u
}
