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

type initialiseListMsg struct{}

func initialiseListCmd() tea.Msg {
	return initialiseListMsg{}
}

type endListUpdateMsg struct {
	itemList []list.Item
	newPath  []string
}

type endFileOpenMsg struct{}

type pickerModel struct {
	list    list.Model
	path    []string
	config  *cfg.Config
	loading bool
}

func (m pickerModel) Init() tea.Cmd {
	return initialiseListCmd
}

func (m pickerModel) changeDir(path string) tea.Cmd {
	m.loading = true

	var newPath []string
	if path == "" {
		newPath = m.path
	} else if path == ".." {
		newPath = m.path[:len(m.path)-1]
	} else {
		newPath = append(m.path, path)
	}

	return func() tea.Msg {
		itemList, err := http.GetItems(m.config, newPath)
		if err != nil {
			return errorMsg{err}
		}

		return endListUpdateMsg{itemList, newPath}
	}
}

func (m pickerModel) openFile(filePath string) tea.Cmd {
	m.loading = true

	return func() tea.Msg {
		err := file.OpenFile(m.config, m.path, filePath)
		if err != nil {
			return errorMsg{err}
		}

		return endFileOpenMsg{}
	}
}

func (m *pickerModel) pickItem(i source.Item) tea.Cmd {
	switch i.ListingType {
	case "dir":
		return m.changeDir(i.Path)

	case "file":
		return m.openFile(i.Path)
	}

	return nil
}

func (m pickerModel) Update(msg tea.Msg) (pickerModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case errorMsg:
		m.list.StopSpinner()
		m.loading = false

	case initialiseListMsg:
		cmds = append(cmds, m.list.StartSpinner(), m.changeDir(""))

	case endListUpdateMsg:
		m.list.StopSpinner()
		m.list.ResetFilter()
		m.list.ResetSelected()
		m.loading = false
		m.list.Title = "/" + strings.Join(msg.newPath, "/")
		m.path = msg.newPath

		cmds = append(cmds, clearErrorCmd, m.list.SetItems(msg.itemList))

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

		switch msg.Type {
		case tea.KeyEnter:
			i, ok := m.list.SelectedItem().(source.Item)
			if ok {
				cmds = append(cmds, m.list.StartSpinner(), m.pickItem(i))
			}
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

func (m pickerModel) View() string {
	view := m.list.View()

	return view
}

func newPicker(config *cfg.Config) pickerModel {
	m := pickerModel{
		list:    list.NewModel([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		config:  config,
		path:    []string{},
		loading: false,
	}

	m.list.Title = "/"
	m.list.SetShowPagination(false)

	return m
}
