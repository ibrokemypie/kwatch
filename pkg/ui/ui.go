package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ibrokemypie/kwatch/pkg/cfg"
)

type childView int

const (
	filePicker childView = iota
	bookmarkPicker
	bookmarkEditor
)

type childModel interface {
	inputFocused() bool
	setSize(int, int)
	Init() tea.Cmd
	Update(tea.Msg) (childModel, tea.Cmd)
	View() string
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}

type mainKeyMap struct {
	ShowFullHelp key.Binding
	HideFullHelp key.Binding
	Quit         key.Binding
	ForceQuit    key.Binding
}

type mainModel struct {
	config       *cfg.Config
	confFilePath string
	currentChild childView
	childModels  []childModel
	helpModel    help.Model
	keys         mainKeyMap
	width        int
	height       int
	err          error
}

func (m mainModel) ShortHelp() []key.Binding {
	bindings := []key.Binding{}

	bindings = append(bindings, m.childModels[m.currentChild].ShortHelp()...)

	if !m.childModels[m.currentChild].inputFocused() {
		bindings = append(bindings, m.keys.Quit)
		bindings = append(bindings, m.keys.ShowFullHelp)
	}

	return bindings
}

func (m mainModel) FullHelp() [][]key.Binding {
	bindings := [][]key.Binding{}

	bindings = append(bindings, m.childModels[m.currentChild].FullHelp()...)

	bindings[len(bindings)-2] = append(bindings[len(bindings)-2], m.keys.HideFullHelp)

	return bindings
}

func (m mainModel) helpView() string {
	return list.DefaultStyles().HelpStyle.Render(m.helpModel.View(m))
}

func (m *mainModel) setSize(width, height int) {
	m.width = width
	m.height = height

	m.updateContents()
}

func (m *mainModel) updateContents() {
	width := m.width

	m.helpModel.Width = width

	availHeight := m.height
	availHeight -= lipgloss.Height(m.helpView())
	availHeight--

	m.childModels[m.currentChild].setSize(width, availHeight)
}

func (m mainModel) Init() tea.Cmd {
	return m.childModels[m.currentChild].Init()
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)

	case errorMsg:
		m.err = msg

	case clearErrorMsg:
		m.err = nil

	case newBookmarkMsg, editBookmarkMsg:
		m.currentChild = bookmarkEditor
		m.updateContents()
		cmds = append(cmds, clearErrorCmd)

	case openBookmarkPickerMsg:
		m.currentChild = bookmarkPicker
		m.updateContents()
		cmds = append(cmds, clearErrorCmd)

	case saveBookmarkMsg:
		err := m.config.WriteConfig(m.confFilePath)
		if err != nil {
			cmds = append(cmds, errorCmd(err))
		}

		m.currentChild = bookmarkPicker
		m.updateContents()
		cmds = append(cmds, clearErrorCmd)

	case updateOpenBookmarkMsg:
		m.currentChild = filePicker
		m.updateContents()
		cmds = append(cmds, clearErrorCmd)

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.ForceQuit) {
			return m, tea.Quit
		}

		if !m.childModels[m.currentChild].inputFocused() {
			switch {
			case key.Matches(msg, m.keys.ShowFullHelp), key.Matches(msg, m.keys.HideFullHelp):
				m.helpModel.ShowAll = !m.helpModel.ShowAll
				m.updateContents()

			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			}
		}
	}

	m.childModels[m.currentChild], cmd = m.childModels[m.currentChild].Update(msg)

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	var view string

	view += m.childModels[m.currentChild].View()
	view += m.helpView()

	if m.err != nil {
		view += fmt.Sprintf("\nAn error occured: %v", m.err)
	}

	return view
}

func NewProgram(config *cfg.Config, confFilePath string) *tea.Program {
	keys := mainKeyMap{
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),

		HideFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),

		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),

		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}

	childModels := []childModel{
		newFilePicker(config),
		newBookmarkPicker(config),
		newBookmarkEditor(config),
	}
	currentChild := bookmarkPicker
	if len(config.Bookmarks) > 0 {
		currentChild = filePicker
	}

	m := mainModel{
		config:       config,
		confFilePath: confFilePath,
		currentChild: currentChild,
		childModels:  childModels,
		helpModel:    help.NewModel(),
		keys:         keys,
		err:          nil,
	}

	p := tea.NewProgram(m, tea.WithMouseCellMotion(), tea.WithAltScreen())
	return p
}
