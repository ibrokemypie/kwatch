package bookmark

import (
	"fmt"
	"net/url"
	"strings"
)

type BackendType int

const (
	HTTP BackendType = iota
)

type Bookmark struct {
	Backend    BackendType
	Address    string
	Path       string
	Username   string
	Password   string
	FileViewer string
}

func (b Bookmark) Title() string {
	return b.Address + b.Path
}

func (b Bookmark) Description() string {
	return string(b.Backend)
}

func (b Bookmark) FilterValue() string {
	return b.Description() + b.Title()
}

func NewBookmark(address *url.URL, path, username, password string) (Bookmark, error) {
	var backend BackendType

	switch address.Scheme {
	case "http", "https":
		backend = HTTP

	default:
		return Bookmark{}, fmt.Errorf("Unsupported backend: %s", address.Scheme)
	}

	address.Path = ""
	address.RawQuery = ""
	address.User = nil
	address.Fragment = ""
	address.Host = strings.TrimSuffix(address.Host, "/")

	path = strings.TrimSuffix(path, "/")

	return Bookmark{
		Backend:    backend,
		Address:    address.String(),
		Path:       path,
		Username:   username,
		Password:   password,
		FileViewer: "mpv",
	}, nil
}
