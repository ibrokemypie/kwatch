package http

import (
	"errors"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/ibrokemypie/kwatch/pkg/source"
	"golang.org/x/net/html"
)

func ParseCaddyList(root *html.Node) ([]list.Item, error) {
	listingNode, err := getListingNode(root)
	if err != nil {
		return nil, err
	}

	listItems := []list.Item{}
	for row := listingNode.FirstChild; row != nil; row = row.NextSibling {
		item, err := extractListing(row)
		if err != nil {
			continue
		} else {
			listItems = append(listItems, item)
		}
	}
	return listItems, nil
}

func getListingNode(node *html.Node) (*html.Node, error) {
	if node.Type == html.ElementNode && node.Data == "tbody" {
		if node.Parent.Parent.Attr[0].Val == "listing" {
			return node, nil
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		listingNode, err := getListingNode(child)
		if err == nil {
			return listingNode, nil
		}
	}
	return nil, errors.New("unable to find the listing table HTML node")
}

func extractListing(node *html.Node) (source.Item, error) {
	item := source.Item{}
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
