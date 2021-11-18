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

type childModel interface {
	inputFocused() bool
	setSize(int, int) childModel
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
	currentChild childModel
	helpModel    help.Model
	keys         mainKeyMap
	width        int
	height       int
	err          error
}

func (m mainModel) ShortHelp() []key.Binding {
	bindings := []key.Binding{}

	bindings = append(bindings, m.currentChild.ShortHelp()...)

	if !m.currentChild.inputFocused() {
		bindings = append(bindings, m.keys.Quit)
		bindings = append(bindings, m.keys.ShowFullHelp)
	}

	return bindings
}

func (m mainModel) FullHelp() [][]key.Binding {
	bindings := [][]key.Binding{}

	bindings = append(bindings, m.currentChild.FullHelp()...)

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

	m.currentChild = m.currentChild.setSize(width, availHeight)
}

func (m mainModel) Init() tea.Cmd {
	return m.currentChild.Init()
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

	case changeViewMsg:
		switch msg.newViewName {
		case "filePicker":
			m.currentChild = newFilePicker(m.config)
		case "bookmarkPicker":
			m.currentChild = newBookmarkPicker(m.config)
		case "bookmarkEditor":
			m.currentChild = newBookmarkEditor(m.config)
		}
		m.updateContents()
		cmds = append(cmds, clearErrorCmd)

	case newBookmarkMsg, editBookmarkMsg:
		m.currentChild = newBookmarkEditor(m.config)
		m.updateContents()

	case saveBookmarkMsg:
		err := m.config.WriteConfig(m.confFilePath)
		if err != nil {
			cmds = append(cmds, errorCmd(err))
		}

		m.currentChild = newBookmarkPicker(m.config)
		m.updateContents()

	case updateOpenBookmarkMsg:
		m.currentChild = newFilePicker(m.config)
		m.updateContents()

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.ForceQuit) {
			return m, tea.Quit
		}

		if !m.currentChild.inputFocused() {
			switch {
			case key.Matches(msg, m.keys.ShowFullHelp), key.Matches(msg, m.keys.HideFullHelp):
				m.helpModel.ShowAll = !m.helpModel.ShowAll
				m.updateContents()

			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			}
		}
	}

	m.currentChild, cmd = m.currentChild.Update(msg)

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	var view string

	view += m.currentChild.View()
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

	m := mainModel{
		config:       config,
		confFilePath: confFilePath,
		helpModel:    help.NewModel(),
		keys:         keys,
		err:          nil,
	}

	if len(m.config.Bookmarks) > 0 {
		m.currentChild = newFilePicker(m.config)
	} else {
		m.currentChild = newBookmarkPicker(m.config)
	}

	p := tea.NewProgram(m, tea.WithMouseCellMotion(), tea.WithAltScreen())
	return p
}
