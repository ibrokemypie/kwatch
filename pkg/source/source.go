package source

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/ibrokemypie/kwatch/pkg/source/bookmark"
	"github.com/ibrokemypie/kwatch/pkg/source/httpSource"
)

type Source interface {
	OpenFile(filePath string) error
	GetItems() ([]list.Item, error)
	ChangeDir(dir string)
	GetPathString() string
	GetAddressString() string
}

func NewSource(b bookmark.Bookmark) Source {
	path := strings.Split(strings.TrimPrefix(b.Path, "/"), "/")

	switch b.Backend {
	case bookmark.HTTP:
		return httpSource.NewHTTPSource(b, path)

	default:
		return nil
	}
}
