package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ibrokemypie/kwatch/pkg/cfg"
)

type bookmarkPickerModel struct {
	config       *cfg.Config
	openBookmark int
	list         list.Model
}

func (m bookmarkPickerModel) Init() tea.Cmd {
	return nil
}

func (m bookmarkPickerModel) Update(msg tea.Msg) (bookmarkPickerModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-1)

	case saveBookmarkMsg:
		bookmarkList := []list.Item{}
		for _, bookmark := range m.config.Bookmarks {
			bookmarkList = append(bookmarkList, bookmark)
		}

		cmds = append(cmds, m.list.SetItems(bookmarkList))

	case updateOpenBookmarkMsg:
		m.openBookmark = msg.newOpenBookmark

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "enter":
			newOpenBookmark := m.list.Index()
			cmds = append(cmds, updateOpenBookmarkCmd(newOpenBookmark))

		case "e":
			if len(m.list.Items()) > 0 {
				cmds = append(cmds, editBookmarkCmd(m.list.Index()))
			}

		case "n":
			cmds = append(cmds, newBookmarkCmd)

		case "b":
			cmds = append(cmds, changeViewCmd(bookmarkPicker))

		case "f":
			cmds = append(cmds, changeViewCmd(filePicker))
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m bookmarkPickerModel) View() string {
	view := m.list.View()

	return view
}

func newBookmarkPicker(config *cfg.Config) bookmarkPickerModel {
	initialList := []list.Item{}
	for _, bookmark := range config.Bookmarks {
		initialList = append(initialList, bookmark)
	}

	m := bookmarkPickerModel{
		config:       config,
		openBookmark: config.DefaultBookmark,
		list:         list.NewModel(initialList, list.NewDefaultDelegate(), 0, 0),
	}

	m.list.Title = "Pick a bookmark to open."
	m.list.SetShowPagination(false)

	return m
}
