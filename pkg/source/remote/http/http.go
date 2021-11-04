package http

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/ibrokemypie/kwatch/pkg/cfg"
	"golang.org/x/net/html"
)

type serverType int

const (
	caddy serverType = iota
	nginx
	apache
)

func GetItems(config *cfg.Config, openBookmark int, path []string) ([]list.Item, error) {
	bookmark := config.Bookmarks[openBookmark]

	address, err := url.Parse(bookmark.Address)
	if err != nil {
		return nil, err
	}

	address.Path = strings.Join(path, "/")

	req, err := http.NewRequest("GET", address.String(), nil)
	if err != nil {
		return nil, err
	}

	if len(bookmark.Username) > 0 {
		req.SetBasicAuth(bookmark.Username, bookmark.Password)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s: %s", req.URL, resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	backend := caddy
	var listItems []list.Item
	switch backend {
	case caddy:
		listItems, err = ParseCaddyList(doc)
		if err != nil {
			return nil, err
		}
	}

	return listItems, nil
}
