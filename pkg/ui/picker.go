package ui

import (
	"errors"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ibrokemypie/kwatch/pkg/cfg"
	"github.com/ibrokemypie/kwatch/pkg/file"
	"github.com/ibrokemypie/kwatch/pkg/source"
	"github.com/ibrokemypie/kwatch/pkg/source/remote/http"
)

type startListUpdateMsg struct{}

func startListUpdate() tea.Msg {
	return startListUpdateMsg{}
}

type endListUpdateMsg struct {
	itemList []list.Item
}

func updateList(cfg *cfg.Config, path []string) tea.Cmd {
	return func() tea.Msg {
		itemList, err := http.GetItems(cfg, path)
		if err != nil {
			return errorMsg{err}
		}

		return endListUpdateMsg{itemList}
	}
}

type pickerModel struct {
	list    list.Model
	path    []string
	config  *cfg.Config
	loading bool
}

func (m pickerModel) Init() tea.Cmd {
	return tea.Batch(
		m.list.StartSpinner(),
		startListUpdate,
	)
}

func pickItem(i source.Item, m *pickerModel) tea.Cmd {
	m.loading = true
	switch i.ListingType {
	case "dir":
		if i.Path == ".." {
			m.path = m.path[:len(m.path)-1]
		} else {
			m.path = append(m.path, i.Path)
		}

		return tea.Batch(m.list.StartSpinner(), updateList(m.config, m.path))

	case "file":
		err := file.OpenFile(m.config, m.path, i.Path)
		if err != nil {
			return errorCmd(err)
		}
	}

	return nil
}

func (m pickerModel) Update(msg tea.Msg) (pickerModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case errorMsg:
		m.list.StopSpinner()
		m.list.NewStatusMessage(msg.err.Error())
		m.loading = false

	case startListUpdateMsg:

		cmds = append(cmds, m.list.StartSpinner(), updateList(m.config, m.path))

	case endListUpdateMsg:
		m.list.StopSpinner()
		m.list.ResetFilter()
		m.list.ResetSelected()
		m.loading = false

		cmds = append(cmds, errorCmd(errors.New("")), m.list.SetItems(msg.itemList))

	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)

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
				cmds = append(cmds, pickItem(i, &m))
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

		case tea.MouseLeft:
			i, ok := m.list.SelectedItem().(source.Item)
			if ok {
				cmds = append(cmds, pickItem(i, &m))
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m pickerModel) View() string {
	return m.list.View()
}

func newPicker(config *cfg.Config, initialPath []string) pickerModel {
	m := pickerModel{
		list:    list.NewModel([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		config:  config,
		path:    initialPath,
		loading: false,
	}

	m.list.Title = "Pick file/directory"
	m.list.SetShowPagination(false)
	m.list.StatusMessageLifetime = 15 * time.Second

	return m
}
