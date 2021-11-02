package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pelletier/go-toml/v2"
)

type item struct {
	listingType string
	name        string
	path        string
}

func (i item) Title() string {
	return i.name
}

func (i item) Description() string {
	return strings.ToTitle(i.listingType)
}

func (i item) FilterValue() string {
	return i.name
}

type Config struct {
	Username   string
	Password   string
	Address    *url.URL
	FileViewer string
}

type model struct {
	list   list.Model
	path   []string
	config *Config
}

type startListUpdateMsg struct{}

type endListUpdateMsg struct {
	itemList []list.Item
}

type errorMsg struct {
	err error
}

func printError(err error) tea.Cmd {
	return func() tea.Msg {
		return errorMsg{err}
	}
}

func updateList(cfg *Config, path []string) tea.Cmd {
	return func() tea.Msg {
		itemList, err := getListings(cfg, path)
		if err != nil {
			return printError(err)
		}

		return endListUpdateMsg{itemList}
	}
}

func startListUpdate() tea.Msg {
	return startListUpdateMsg{}
}

func selectItem(i item, m *model) tea.Cmd {
	switch i.listingType {
	case "dir":
		if i.path == ".." {
			m.path = m.path[:len(m.path)-1]
		} else {
			m.path = append(m.path, i.path)
		}

		return tea.Batch(m.list.StartSpinner(), updateList(m.config, m.path))

	case "file":
		err := openFile(m.config, m.path, i.path)
		if err != nil {
			return printError(err)
		}
	}

	return nil
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.list.StartSpinner(),
		startListUpdate,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case errorMsg:
		m.list.StopSpinner()
		cmds = append(cmds, m.list.NewStatusMessage(msg.err.Error()))

	case startListUpdateMsg:
		cmds = append(cmds, m.list.StartSpinner(), updateList(m.config, m.path))

	case endListUpdateMsg:
		m.list.StopSpinner()
		m.list.ResetFilter()
		m.list.ResetSelected()
		cmds = append(cmds, m.list.SetItems(msg.itemList))

	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.Type {
		case tea.KeyEnter:
			i, ok := m.list.SelectedItem().(item)
			if ok {
				cmds = append(cmds, selectItem(i, &m))
			}
		}

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			m.list.CursorUp()

		case tea.MouseWheelDown:
			m.list.CursorDown()

		case tea.MouseLeft:
			i, ok := m.list.SelectedItem().(item)
			if ok {
				cmds = append(cmds, selectItem(i, &m))
			}
		}

	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return m.list.View()
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
	addressString := flag.String("a", "", "Root caddy fileserver address [required]")
	fileViewer := flag.String("o", "", "Program to open files with [optional] [defaults to mpv]")
	confFile := flag.String("c", confDir+"/kwatch.toml", "Configuration file [optional]")
	writeToConfig := flag.Bool("w", false, "Write current arguments to config file [optional]")
	flag.Parse()

	var address *url.URL
	if len(*addressString) > 0 {
		address, err = url.Parse(*addressString)
		if err != nil {
			log.Fatal(err)
		}
		if len(address.Scheme) <= 0 {
			fmt.Println("Address requires scheme (http/https)")
			os.Exit(1)
		}
	}

	cfg := Config{*username, *password, address, *fileViewer}

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

	if cfg.Address == nil {
		flag.Usage()
		os.Exit(1)
	}

	initialPath := strings.Split(cfg.Address.Path, "/")
	cfg.Address.Path = ""

	if len(cfg.FileViewer) <= 0 {
		cfg.FileViewer = "mpv"
	}

	_, err = exec.LookPath(cfg.FileViewer)
	if err != nil {
		log.Fatal(err)
	}

	m := model{
		list:   list.NewModel([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		config: &cfg,
		path:   initialPath,
	}

	m.list.Title = "Pick file/directory"
	m.list.SetShowPagination(false)
	m.list.StatusMessageLifetime = 5 * time.Second

	p := tea.NewProgram(m, tea.WithMouseCellMotion())
	p.EnterAltScreen()

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}

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
	if cfg.Address == nil {
		cfg.Address = readCfg.Address
	}
	if len(cfg.FileViewer) <= 0 {
		cfg.FileViewer = readCfg.FileViewer
	}

	return nil
}

func getListings(cfg *Config, path []string) ([]list.Item, error) {
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

	listItems, err := parseList(doc)
	if err != nil {
		return nil, err
	}

	return listItems, nil
}

func openFile(cfg *Config, path []string, filePath string) error {
	addressCopy := cfg.Address

	addressCopy.User = url.UserPassword(cfg.Username, cfg.Password)
	addressCopy.Path = strings.Join(path, "/") + filePath
	runCMD := exec.Command(cfg.FileViewer, addressCopy.String())

	return runCMD.Run()
}

func parseList(root *html.Node) ([]list.Item, error) {
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

func extractListing(node *html.Node) (item, error) {
	item := item{}
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

	return item, errors.New("no listitem could be extracted")
}
