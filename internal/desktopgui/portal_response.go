package desktopgui

import (
	"errors"
	"fmt"
	"net/url"
)

func selectedPathFromPortalURIs(uris []string) (string, error) {
	if len(uris) == 0 {
		return "", errors.New("portal response did not include a selected file")
	}
	return filePathFromURI(uris[0])
}

func filePathFromURI(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse file URI: %w", err)
	}
	if parsed.Scheme != "file" {
		return "", fmt.Errorf("unsupported URI scheme %q", parsed.Scheme)
	}
	if parsed.Host != "" && parsed.Host != "localhost" {
		return "", fmt.Errorf("unsupported file URI host %q", parsed.Host)
	}
	return url.PathUnescape(parsed.Path)
}
