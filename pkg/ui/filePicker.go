package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ibrokemypie/kwatch/pkg/cfg"
	"github.com/ibrokemypie/kwatch/pkg/source"
	"github.com/ibrokemypie/kwatch/pkg/source/sourceItem"
)

type filePickerKeymap struct {
	SelectFile         key.Binding
	GoUp               key.Binding
	ShowBookmarkPicker key.Binding
}

type filePickerModel struct {
	config        *cfg.Config
	currentSource source.Source
	list          list.Model
	loading       bool
	keys          filePickerKeymap
}

func (m filePickerModel) ShortHelp() []key.Binding {
	bindings := []key.Binding{}

	if len(m.list.Items()) > 0 {
		bindings = append(bindings, m.keys.SelectFile)
	}
	bindings = append(bindings, m.list.ShortHelp()...)

	return bindings
}

func (m filePickerModel) FullHelp() [][]key.Binding {
	bindings := m.list.FullHelp()

	bindings[1] = append(bindings[1], m.keys.SelectFile, m.keys.GoUp, m.keys.ShowBookmarkPicker)

	return bindings
}

func (m *filePickerModel) setSize(width, height int) {
	m.list.SetSize(width, height)
}

func (m filePickerModel) inputFocused() bool {
	filterState := m.list.FilterState()

	switch filterState {
	case list.Filtering:
		return true

	default:
		return false
	}
}

func (m filePickerModel) Init() tea.Cmd {
	if m.config.GetDefaultBookmark() != -1 {
		return updateOpenBookmarkCmd(m.config.GetDefaultBookmark())
	}

	return nil
}

func (m filePickerModel) changeDir(path string) tea.Cmd {
	m.currentSource.ChangeDir(path)

	return func() tea.Msg {
		itemList, err := m.currentSource.GetItems()
		if err != nil {
			return errorMsg{err}
		}

		return endListUpdateMsg{itemList}
	}
}

func (m filePickerModel) initialiseFileList() tea.Cmd {
	return func() tea.Msg {
		itemList, err := m.currentSource.GetItems()
		if err != nil {
			return errorMsg{err}
		}

		return endListUpdateMsg{itemList}
	}
}

func (m filePickerModel) openFile(filePath string) tea.Cmd {
	return func() tea.Msg {
		err := m.currentSource.OpenFile(filePath)
		if err != nil {
			return errorMsg{err}
		}

		return endFileOpenMsg{}
	}
}

func (m filePickerModel) pickItem(i sourceItem.Item) tea.Cmd {
	switch i.ListingType {
	case "dir":
		return m.changeDir(i.Path)

	case "file":
		return m.openFile(i.Path)
	}

	return nil
}

func (m filePickerModel) Update(msg tea.Msg) (childModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case errorMsg:
		m.list.StopSpinner()
		m.loading = false

	case updateOpenBookmarkMsg:
		m.loading = true
		m.list.SetItems([]list.Item{})

		bookmark := m.config.GetBookmark(msg.newOpenBookmark)
		m.currentSource = source.NewSource(bookmark)

		pathString := bookmark.Path
		m.list.Title = bookmark.Address + pathString

		cmds = append(cmds, m.list.StartSpinner(), m.initialiseFileList())

	case endListUpdateMsg:
		m.list.StopSpinner()
		m.list.ResetFilter()
		m.list.ResetSelected()
		m.loading = false
		m.list.Title = m.currentSource.GetAddressString() + "/" + m.currentSource.GetPathString()

		cmds = append(cmds, clearErrorCmd, m.list.SetItems(msg.itemList))

	case endFileOpenMsg:
		m.list.StopSpinner()
		m.loading = false
		cmds = append(cmds, clearErrorCmd)

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		if m.loading {
			break
		}

		switch {
		case key.Matches(msg, m.keys.SelectFile):
			i, ok := m.list.SelectedItem().(sourceItem.Item)
			if ok {
				m.loading = true
				cmds = append(cmds, m.list.StartSpinner(), m.pickItem(i))
			}

		case key.Matches(msg, m.keys.GoUp):
			for _, item := range m.list.Items() {
				sourceItem := item.(sourceItem.Item)
				if sourceItem.Path == ".." {
					m.loading = true
					cmds = append(cmds, m.list.StartSpinner(), m.pickItem(sourceItem))
				}
			}

		case key.Matches(msg, m.keys.ShowBookmarkPicker):
			cmds = append(cmds, openBookmarkPickerCmd)
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
	return &m, tea.Batch(cmds...)
}

func (m filePickerModel) View() string {
	view := m.list.View()

	return view
}

func newFilePicker(config *cfg.Config) *filePickerModel {
	listModel := list.NewModel([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	listModel.SetShowPagination(false)
	listModel.SetShowHelp(false)
	listModel.DisableQuitKeybindings()

	listModel.KeyMap.ShowFullHelp.SetEnabled(false)
	listModel.KeyMap.CloseFullHelp.SetEnabled(false)

	keys := filePickerKeymap{
		SelectFile: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		GoUp: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "up directory"),
		),
		ShowBookmarkPicker: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "bookmarks"),
		),
	}

	m := filePickerModel{
		config:  config,
		list:    listModel,
		loading: false,
		keys:    keys,
	}

	return &m
}
