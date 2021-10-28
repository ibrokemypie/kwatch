package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type listItem struct {
	listingType string
	name        string
	path        string
}

var (
	username string
	password string
	address  string
)

func init() {
	flag.StringVar(&username, "u", "", "HTTP Username [optional]")
	flag.StringVar(&password, "p", "", "HTTP Password [optional]")
	flag.StringVar(&address, "a", "", "Root caddy fileserver address [required]")
}

func main() {
	flag.Parse()

	if len(address) <= 0 {
		flag.Usage()
		os.Exit(1)
	}

	pickItem([]string{})
}

func getListings(path []string) ([]listItem, error) {
	var pathString string
	for _, v := range path {
		pathString += v
	}
	getURL, err := url.Parse(address + pathString)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", getURL.String(), nil)
	if err != nil {
		return nil, err
	}

	if len(username) > 0 {
		req.SetBasicAuth(username, password)
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

	listItems, err := parseList(doc)
	if err != nil {
		return nil, err
	}

	return listItems, nil
}

func pickItem(currPath []string) {
	listings, err := getListings(currPath)
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range listings {
		fmt.Printf("%d: %s\t%s\n", k, v.name, v.listingType)
	}

	pickNo := -1
	scanner := bufio.NewScanner(os.Stdin)
	for pickNo == -1 {
		fmt.Printf("Enter choice [%d-%d]:", 0, len(listings)-1)
		scanner.Scan()
		pickStr := scanner.Text()

		pickNo, err = strconv.Atoi(pickStr)
		if err != nil {
			log.Println(err)
			continue
		}
		if pickNo < 0 || pickNo >= len(listings) {
			log.Printf("Choice must be between %d and %d", 0, len(listings)-1)
			pickNo = -1
			continue
		}
	}

	pick := listings[pickNo]
	if pick.listingType == "dir" {
		if pick.path != ".." {
			currPath = append(currPath, pick.path)
		} else {
			currPath = currPath[:len(currPath)-1]
		}
	} else {
		playMpv(currPath, pick.path)
	}

	pickItem(currPath)
}

func playMpv(currPath []string, filePath string) error {
	var pathString string
	for _, v := range currPath {
		pathString += v
	}

	fileURL, err := url.Parse(address + pathString + filePath)
	if err != nil {
		return err
	}
	fileURL.User = url.UserPassword(username, password)
	mpvCMD := exec.Command("mpv", fileURL.String())
	fmt.Println(mpvCMD)

	return mpvCMD.Run()
}

func parseList(root *html.Node) ([]listItem, error) {
	listingNode, err := getListingNode(root)
	if err != nil {
		return nil, err
	}

	listItems := []listItem{}
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
	return nil, errors.New("Unable to find the listing table HTML node")
}

func extractListing(node *html.Node) (listItem, error) {
	for col := node.FirstChild; col != nil; col = col.NextSibling {
		for child := col.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode && child.Data == "a" {
				var item listItem
				for _, attr := range child.Attr {
					if attr.Key == "href" {
						item.path = strings.TrimPrefix(attr.Val, ".")
					}
				}

				for linkChild := child.FirstChild; linkChild != nil; linkChild = linkChild.NextSibling {
					if linkChild.Type == html.ElementNode && linkChild.Data == "span" {
						item.name = linkChild.FirstChild.Data
					}
				}

				if strings.HasSuffix(item.path, "/") || item.path == ".." {
					item.listingType = "dir"
				} else {
					item.listingType = "file"
				}

				return item, nil
			}
		}

	}

	return listItem{}, errors.New("No listitem could be extracted")
}
