package httpSource

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/ibrokemypie/kwatch/pkg/source/bookmark"
	"github.com/ibrokemypie/kwatch/pkg/source/sourceItem"
	"golang.org/x/net/html"
)

type Backend struct {
	bookmark    bookmark.Bookmark
	currentPath []string
}

func NewHTTPSource(bookmark bookmark.Bookmark, path []string) *Backend {
	return &Backend{bookmark, path}
}

func (b Backend) OpenFile(filePath string) error {
	address, err := url.Parse(b.bookmark.Address)
	if err != nil {
		return err
	}

	address.User = url.UserPassword(b.bookmark.Username, b.bookmark.Password)
	address.Path = b.GetPathString() + "/" + filePath
	runCMD := exec.Command(b.bookmark.FileViewer, address.String())

	err = runCMD.Run()
	if err != nil {
		return fmt.Errorf("%s: %s", runCMD.String(), err.Error())
	}

	return nil
}

func (b *Backend) ChangeDir(dir string) {
	if dir == ".." {
		b.currentPath = b.currentPath[:len(b.currentPath)-1]
	} else {
		b.currentPath = append(b.currentPath, dir)
	}
}

func (b Backend) GetItems() ([]list.Item, error) {
	address, err := url.Parse(b.bookmark.Address)
	if err != nil {
		return nil, err
	}

	address.Path = b.GetPathString()

	req, err := http.NewRequest("GET", address.String(), nil)
	if err != nil {
		return nil, err
	}

	if len(b.bookmark.Username) > 0 {
		req.SetBasicAuth(b.bookmark.Username, b.bookmark.Password)
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

	var listItems []list.Item

	listItems, err = parseCaddyList(doc)
	if err != nil {
		return nil, err
	}

	return listItems, nil
}

func (b Backend) GetPathString() string {
	ioutil.WriteFile("output.txt", []byte(fmt.Sprint(b.currentPath)), 0644)
	return strings.Join(b.currentPath, "/")
}

func (b Backend) GetAddressString() string {
	return b.bookmark.Address
}

func parseCaddyList(root *html.Node) ([]list.Item, error) {
	listingNode, err := getCaddyListingNode(root)
	if err != nil {
		return nil, err
	}

	listItems := []list.Item{}
	for row := listingNode.FirstChild; row != nil; row = row.NextSibling {
		item, err := extractCaddyListing(row)
		if err != nil {
			continue
		} else {
			listItems = append(listItems, item)
		}
	}
	return listItems, nil
}

func getCaddyListingNode(node *html.Node) (*html.Node, error) {
	if node.Type == html.ElementNode && node.Data == "tbody" {
		if node.Parent.Parent.Attr[0].Val == "listing" {
			return node, nil
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		listingNode, err := getCaddyListingNode(child)
		if err == nil {
			return listingNode, nil
		}
	}
	return nil, errors.New("unable to find the listing table HTML node")
}

func extractCaddyListing(node *html.Node) (sourceItem.Item, error) {
	item := sourceItem.Item{}
	for col := node.FirstChild; col != nil; col = col.NextSibling {
		for child := col.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode && child.Data == "a" {
				for _, attr := range child.Attr {
					if attr.Key == "href" {
						item.Path = attr.Val
					}
				}

				for linkChild := child.FirstChild; linkChild != nil; linkChild = linkChild.NextSibling {
					if linkChild.Type == html.ElementNode && linkChild.Data == "span" {
						item.Name = linkChild.FirstChild.Data
					}
				}

				if strings.HasSuffix(item.Path, "/") || item.Path == ".." {
					item.ListingType = "dir"
				} else {
					item.ListingType = "file"
				}

				if item.Path != ".." {
					item.Path = strings.Replace(item.Path, "/", "", -1)
					item.Path = strings.TrimPrefix(item.Path, ".")
					cleanPath, err := url.PathUnescape(item.Path)
					if err != nil {
						return item, err
					}
					item.Path = cleanPath
				}

				return item, nil
			}
		}

	}

	return item, errors.New("no listitem could be extracted")
}
