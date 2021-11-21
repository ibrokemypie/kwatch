package sourceItem

import (
	"strings"
)

type Item struct {
	ListingType string
	Name        string
	Path        string
}

func (i Item) Title() string {
	return i.Name
}

func (i Item) Description() string {
	return strings.ToTitle(i.ListingType)
}

func (i Item) FilterValue() string {
	return i.Name
}
