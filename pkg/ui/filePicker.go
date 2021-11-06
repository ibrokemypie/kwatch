package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ibrokemypie/kwatch/pkg/cfg"
	"github.com/ibrokemypie/kwatch/pkg/file"
	"github.com/ibrokemypie/kwatch/pkg/source"
	"github.com/ibrokemypie/kwatch/pkg/source/remote/http"
)

type filePickerModel struct {
	config       *cfg.Config
	openBookmark int
	currentPath  []string
	list         list.Model
	loading      bool
}

func (m filePickerModel) Init() tea.Cmd {
	return initialiseListCmd
}

func (m *filePickerModel) changeDir(path string) tea.Cmd {
	m.loading = true

	var newPath []string
	if path == "" {
		newPath = m.currentPath
	} else if path == ".." {
		newPath = m.currentPath[:len(m.currentPath)-1]
	} else {
		newPath = append(m.currentPath, path)
	}

	return func() tea.Msg {
		itemList, err := http.GetItems(m.config, m.openBookmark, newPath)
		if err != nil {
			return errorMsg{err}
		}

		return endListUpdateMsg{itemList, newPath}
	}
}

func (m *filePickerModel) initialiseListItems() tea.Cmd {
	return m.changeDir("")
}

func (m filePickerModel) openFile(filePath string) tea.Cmd {
	m.loading = true

	return func() tea.Msg {
		err := file.OpenFile(m.config, m.openBookmark, m.currentPath, filePath)
		if err != nil {
			return errorMsg{err}
		}

		return endFileOpenMsg{}
	}
}

func (m *filePickerModel) pickItem(i source.Item) tea.Cmd {
	switch i.ListingType {
	case "dir":
		return m.changeDir(i.Path)

	case "file":
		return m.openFile(i.Path)
	}

	return nil
}

func (m filePickerModel) Update(msg tea.Msg) (filePickerModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case errorMsg:
		m.list.StopSpinner()
		m.loading = false

	case initialiseListMsg:
		cmds = append(cmds, m.list.StartSpinner(), m.initialiseListItems())

	case endListUpdateMsg:
		m.list.StopSpinner()
		m.list.ResetFilter()
		m.list.ResetSelected()
		m.loading = false
		m.list.Title = m.config.Bookmarks[m.openBookmark].Address + "/" + strings.Join(msg.newPath, "/")
		m.currentPath = msg.newPath

		cmds = append(cmds, clearErrorCmd, m.list.SetItems(msg.itemList))

	case updateOpenBookmarkMsg:
		m.openBookmark = msg.newOpenBookmark
		bookmark := m.config.Bookmarks[m.openBookmark]
		pathString := bookmark.Path
		m.list.Title = bookmark.Address + pathString

		m.currentPath = strings.Split(strings.TrimPrefix(pathString, "/"), "/")

		cmds = append(cmds, initialiseListCmd)

	case endFileOpenMsg:
		m.list.StopSpinner()
		m.loading = false
		cmds = append(cmds, clearErrorCmd)

	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-1)

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		if m.loading {
			break
		}

		switch msg.String() {
		case "enter":
			i, ok := m.list.SelectedItem().(source.Item)
			if ok {
				cmds = append(cmds, m.list.StartSpinner(), m.pickItem(i))
			}

		case "b":
			cmds = append(cmds, changeViewCmd(bookmarkPicker))

		case "f":
			cmds = append(cmds, changeViewCmd(filePicker))
		}

	case tea.MouseMsg:
		if m.loading {
			break
		}

		switch msg.Type {
		case tea.MouseWheelUp:
			m.list.CursorUp()

		case tea.MouseWheelDown:
			m.list.CursorDown()

		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m filePickerModel) View() string {
	view := m.list.View()

	return view
}

func newFilePicker(config *cfg.Config) filePickerModel {
	defaultBookmark := source.Bookmark{}
	initialPath := []string{}

	if len(config.Bookmarks) > 0 {
		defaultBookmark = config.Bookmarks[config.DefaultBookmark]

		pathString := defaultBookmark.Path
		initialPath = strings.Split(strings.TrimPrefix(pathString, "/"), "/")
	}

	m := filePickerModel{
		config:       config,
		openBookmark: config.DefaultBookmark,
		currentPath:  initialPath,
		list:         list.NewModel([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		loading:      false,
	}

	m.list.Title = defaultBookmark.Title()
	m.list.SetShowPagination(false)

	return m
}
