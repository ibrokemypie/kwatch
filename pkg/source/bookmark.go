package source

import (
	"fmt"
	"net/url"
	"strings"
)

type Bookmark struct {
	Backend  string
	Address  string
	Path     string
	Username string
	Password string
}

func (b Bookmark) Title() string {
	return b.Address + b.Path
}

func (b Bookmark) Description() string {
	return b.Backend
}

func (b Bookmark) FilterValue() string {
	return b.Description() + b.Title()
}

func NewBookmark(address *url.URL, path, username, password string) (Bookmark, error) {
	var backend string

	switch address.Scheme {
	case "http", "https":
		backend = "caddy"

	default:
		return Bookmark{}, fmt.Errorf("Unsupported backend: %s", address.Scheme)
	}

	address.Path = ""
	address.RawQuery = ""
	address.User = nil
	address.Fragment = ""
	address.Host = strings.TrimSuffix(address.Host, "/")

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = strings.TrimSuffix(path, "/")

	return Bookmark{
		Backend:  backend,
		Address:  address.String(),
		Path:     path,
		Username: username,
		Password: password,
	}, nil
}
