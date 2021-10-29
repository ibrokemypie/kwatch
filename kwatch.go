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

	"github.com/pelletier/go-toml/v2"
)

type listItem struct {
	listingType string
	name        string
	path        string
}

type Config struct {
	Username      string
	Password      string
	AddressString string
	AddressURL    *url.URL
	FileViewer    string
}

func main() {
	confDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Unable to find user config dir: %s", err)
		confDir, err = os.Getwd()
		if err != nil {
			log.Fatalf("Unable to find current working directory dir: %s", err)
		}
	}
	username := flag.String("u", "", "HTTP Username [optional]")
	password := flag.String("p", "", "HTTP Password [optional]")
	address := flag.String("a", "", "Root caddy fileserver address [required]")
	fileViewer := flag.String("o", "", "Program to open files with [optional] [defaults to mpv]")
	confFile := flag.String("c", confDir+"/kwatch.toml", "Configuration file [optional]")
	writeToConfig := flag.Bool("w", false, "Write current arguments to config file [optional]")
	flag.Parse()

	cfg := Config{*username, *password, *address, nil, *fileViewer}

	if *writeToConfig {
		fmt.Println("Attempting to write current options to config file " + *confFile)
		err = writeConfig(&cfg, &confDir, confFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = readConfig(&cfg, confFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}

	if len(cfg.AddressString) <= 0 {
		flag.Usage()
		os.Exit(1)
	}

	if len(cfg.FileViewer) <= 0 {
		cfg.FileViewer = "mpv"
	}

	addressURL, err := url.Parse(cfg.AddressString)
	if err != nil {
		log.Fatal(err)
	}
	cfg.AddressURL = addressURL

	if len(cfg.AddressURL.Scheme) <= 0 {
		cfg.AddressURL.Scheme = "https"
		log.Println("No HTTP scheme in provided address, Trying HTTPS.")
	}

	cfg.AddressURL.Path = strings.TrimSuffix(addressURL.Path, "/")

	_, err = exec.LookPath(cfg.FileViewer)
	if err != nil {
		log.Fatal(err)
	}

	pickItem(&cfg)
}

func writeConfig(cfg *Config, confDir, confFile *string) error {
	bytes, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.MkdirAll(*confDir, os.ModeAppend)
	if err != nil {
		return err
	}

	file, err := os.Create(*confFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func readConfig(cfg *Config, confFile *string) error {
	var readCfg Config
	bytes, err := os.ReadFile(*confFile)
	if err != nil {
		return err
	}
	err = toml.Unmarshal(bytes, &readCfg)
	if err != nil {
		return err
	}

	if len(cfg.Username) <= 0 {
		cfg.Username = readCfg.Username
	}
	if len(cfg.Password) <= 0 {
		cfg.Password = readCfg.Password
	}
	if len(cfg.AddressString) <= 0 {
		cfg.AddressString = readCfg.AddressString
	}
	if len(cfg.FileViewer) <= 0 {
		cfg.FileViewer = readCfg.FileViewer
	}

	return nil
}

func pickItem(cfg *Config) {
	listings, err := getListings(cfg)
	if err != nil {
		log.Fatal(err)
	}

	printListings(listings)

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
			printListings(listings)
			continue
		}

		pickNo, err = strconv.Atoi(pickStr)
		if err != nil {
			log.Println(err)
			pickNo = -1
			continue
		}

		if pickNo < 0 || pickNo >= len(listings) {
			fmt.Printf("Choice must be between %d and %d\n", 0, len(listings)-1)
			pickNo = -1
			continue
		}
	}

	pick := listings[pickNo]
	switch pick.listingType {
	case "dir":
		if pick.path == ".." {
			oldPath := cfg.AddressURL.Path
			oldPathSlice := strings.Split(oldPath, "/")
			newPathSlice := oldPathSlice[:len(oldPathSlice)-1]
			newPath := strings.Join(newPathSlice, "/")
			cfg.AddressURL.Path = newPath
		} else {
			cfg.AddressURL.Path = cfg.AddressURL.Path + pick.path
		}
	case "file":
		err = openFile(cfg, pick.path)
		if err != nil {
			log.Printf("Error opening file with %s: %s\n", cfg.FileViewer, err)
		}
	}

	pickItem(cfg)
}

func printListings(listings []*listItem) {
	for k, v := range listings {
		fmt.Printf("%d: %s\n", k, v.name)
	}
}

func getListings(cfg *Config) ([]*listItem, error) {
	req, err := http.NewRequest("GET", cfg.AddressURL.String(), nil)
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

	listItems, err := parseList(doc)
	if err != nil {
		return nil, err
	}

	return listItems, nil
}

func openFile(cfg *Config, filePath string) error {
	cfg.AddressURL.User = url.UserPassword(cfg.Username, cfg.Password)
	cfg.AddressURL.Path = cfg.AddressURL.Path + filePath
	runCMD := exec.Command(cfg.FileViewer, cfg.AddressURL.String())
	runCMD.Stderr = os.Stderr
	fmt.Println(runCMD)

	return runCMD.Run()
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
