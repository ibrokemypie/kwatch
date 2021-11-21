package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ibrokemypie/kwatch/pkg/cfg"
)

type bookmarkPickerKeymap struct {
	NewBookmark    key.Binding
	EditBookmark   key.Binding
	SelectBookmark key.Binding
	ShowFilePicker key.Binding
}

type bookmarkPickerModel struct {
	config       *cfg.Config
	openBookmark int
	list         list.Model
	keys         bookmarkPickerKeymap
}

func (m bookmarkPickerModel) ShortHelp() []key.Binding {
	bindings := []key.Binding{}

	if len(m.list.Items()) > 0 {
		bindings = append(bindings, m.keys.SelectBookmark, m.keys.EditBookmark)
	}
	bindings = append(bindings, m.keys.NewBookmark)
	bindings = append(bindings, m.list.ShortHelp()...)

	return bindings
}

func (m bookmarkPickerModel) FullHelp() [][]key.Binding {
	bindings := m.list.FullHelp()

	bindings[1] = append(bindings[1], m.keys.NewBookmark, m.keys.EditBookmark, m.keys.SelectBookmark)

	return bindings
}

func (m *bookmarkPickerModel) setSize(width, height int) {
	m.list.SetSize(width, height)
}

func (m bookmarkPickerModel) inputFocused() bool {
	filterState := m.list.FilterState()

	switch filterState {
	case list.Filtering:
		return true

	default:
		return false
	}
}

func (m bookmarkPickerModel) Init() tea.Cmd {
	return nil
}

func (m bookmarkPickerModel) Update(msg tea.Msg) (childModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
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

		switch {
		case key.Matches(msg, m.keys.SelectBookmark):
			newOpenBookmark := m.list.Index()
			cmds = append(cmds, updateOpenBookmarkCmd(newOpenBookmark))

		case key.Matches(msg, m.keys.EditBookmark):
			if len(m.list.Items()) > 0 {
				cmds = append(cmds, editBookmarkCmd(m.list.Index()))
			}

		case key.Matches(msg, m.keys.NewBookmark):
			cmds = append(cmds, newBookmarkCmd)

		case key.Matches(msg, m.keys.ShowFilePicker):
			cmds = append(cmds, openBookmarkPickerCmd)
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return &m, tea.Batch(cmds...)
}

func (m bookmarkPickerModel) View() string {
	view := m.list.View()

	return view
}

func newBookmarkPicker(config *cfg.Config) *bookmarkPickerModel {
	initialList := []list.Item{}
	for _, bookmark := range config.Bookmarks {
		initialList = append(initialList, bookmark)
	}

	listModel := list.NewModel(initialList, list.NewDefaultDelegate(), 0, 0)
	listModel.SetShowPagination(false)
	listModel.SetShowHelp(false)
	listModel.DisableQuitKeybindings()

	listModel.KeyMap.ShowFullHelp.SetEnabled(false)
	listModel.KeyMap.CloseFullHelp.SetEnabled(false)

	listModel.Title = "Pick a bookmark to open."

	keys := bookmarkPickerKeymap{
		NewBookmark: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),

		EditBookmark: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),

		SelectBookmark: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),

		ShowFilePicker: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "files"),
		),
	}

	m := bookmarkPickerModel{
		config:       config,
		openBookmark: config.DefaultBookmark,
		list:         listModel,
		keys:         keys,
	}

	return &m
}
