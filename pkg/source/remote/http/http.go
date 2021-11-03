package http

import (
	"fmt"
	"net/http"
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

func GetItems(cfg *cfg.Config, path []string) ([]list.Item, error) {
	req, err := http.NewRequest("GET", cfg.Address.String()+strings.Join(path, "/"), nil)
	if err != nil {
		return nil, err
	}

	if len(cfg.Username) > 0 {
		req.SetBasicAuth(cfg.Username, cfg.Password)
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
