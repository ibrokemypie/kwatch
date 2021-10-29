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

func main() {
	username := flag.String("u", "", "HTTP Username [optional]")
	password := flag.String("p", "", "HTTP Password [optional]")
	address := flag.String("a", "", "Root caddy fileserver address [required]")
	flag.Parse()

	if len(*address) <= 0 {
		flag.Usage()
		os.Exit(1)
	}

	addressURL, err := url.Parse(*address)
	if err != nil {
		log.Fatal(err)
	}

	if len(addressURL.Scheme) <= 0 {
		addressURL.Scheme = "https"
		fmt.Println("No HTTP scheme in provided address, Trying HTTPS.")
	}

	if addressURL.Path == "/" {
		addressURL.Path = ""
	}

	pickItem(addressURL, username, password)
}

func pickItem(addressURL *url.URL, username, password *string) {
	listings, err := getListings(addressURL, username, password)
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range listings {
		fmt.Printf("%d: %s\n", k, v.name)
	}

	pickNo := -1
	scanner := bufio.NewScanner(os.Stdin)
	for pickNo == -1 {
		fmt.Printf("Enter choice [%d-%d, p, q]:", 0, len(listings)-1)
		scanner.Scan()
		pickStr := scanner.Text()

		if strings.ToLower(pickStr) == "q" {
			os.Exit(0)
		}

		if strings.ToLower(pickStr) == "p" {
			for k, v := range listings {
				fmt.Printf("%d: %s\n", k, v.name)
			}
			continue
		}

		pickNo, err = strconv.Atoi(pickStr)
		if err != nil {
			log.Println(err)
			pickNo = -1
			continue
		}

		if pickNo < 0 || pickNo >= len(listings) {
			log.Printf("Choice must be between %d and %d", 0, len(listings)-1)
			pickNo = -1
			continue
		}
	}

	pick := listings[pickNo]
	switch pick.listingType {
	case "dir":
		if pick.path == ".." {
			oldPath := addressURL.Path
			oldPathSlice := strings.Split(oldPath, "/")
			newPathSlice := oldPathSlice[:len(oldPathSlice)-1]
			newPath := strings.Join(newPathSlice, "/")
			addressURL.Path = newPath
		} else {
			addressURL.Path = addressURL.Path + pick.path
		}
	case "file":
		playMpv(*addressURL, pick.path, *username, *password)
	}

	pickItem(addressURL, username, password)
}

func getListings(addressURL *url.URL, username, password *string) ([]*listItem, error) {
	req, err := http.NewRequest("GET", addressURL.String(), nil)
	if err != nil {
		return nil, err
	}

	if len(*username) > 0 {
		req.SetBasicAuth(*username, *password)
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

func playMpv(addressURL url.URL, filePath string, username, password string) {
	addressURL.User = url.UserPassword(username, password)
	addressURL.Path = addressURL.Path + filePath
	mpvCMD := exec.Command("mpv", addressURL.String())
	fmt.Println(mpvCMD)

	err := mpvCMD.Run()
	if err != nil {
		fmt.Printf("Error opening file in MPV: %s\n", err)
	}
}

func parseList(root *html.Node) ([]*listItem, error) {
	listingNode, err := getListingNode(root)
	if err != nil {
		return nil, err
	}

	listItems := []*listItem{}
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

func extractListing(node *html.Node) (*listItem, error) {
	item := new(listItem)
	for col := node.FirstChild; col != nil; col = col.NextSibling {
		for child := col.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode && child.Data == "a" {
				for _, attr := range child.Attr {
					if attr.Key == "href" {
						path, err := url.PathUnescape(attr.Val)
						if err != nil {
							return item, err
						}
						item.path = path
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

				if item.path != ".." {
					item.path = strings.TrimPrefix(item.path, ".")
					item.path = strings.TrimSuffix(item.path, "/")
				}

				return item, nil
			}
		}

	}

	return item, errors.New("No listitem could be extracted")
}
