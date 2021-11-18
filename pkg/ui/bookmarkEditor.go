package ui

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ibrokemypie/kwatch/pkg/cfg"
	"github.com/ibrokemypie/kwatch/pkg/source"
)

type bookmarkEditorKeymap struct {
	NextField   key.Binding
	PrevField   key.Binding
	Select      key.Binding
	LeaveEditor key.Binding
}

var (
	blurredStyle       = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})
	focusedStyle       = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"})
	cursorStyle        = focusedStyle.Copy()
	helpStyle          = blurredStyle.Copy()
	focusedButtonStyle = focusedStyle.Copy().Padding(0, 1)
	blurredButtonStyle = blurredStyle.Copy().Padding(0, 1)
	titleBarStyle      = lipgloss.NewStyle().Padding(0, 0, 1, 2)
	titleStyle         = lipgloss.NewStyle().
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("230")).
				Padding(0, 1)
)

type bookmarkEditorModel struct {
	config        *cfg.Config
	bookmarkIndex int
	createNew     bool
	inputs        []textinput.Model
	inputCount    int
	focusIndex    int
	width         int
	height        int
	keys          bookmarkEditorKeymap
}

func (m bookmarkEditorModel) ShortHelp() []key.Binding {
	bindings := []key.Binding{}

	bindings = append(bindings, m.keys.Select, m.keys.NextField, m.keys.PrevField, m.keys.LeaveEditor)

	return bindings
}

func (m bookmarkEditorModel) FullHelp() [][]key.Binding {
	bindings := [][]key.Binding{}

	bindings[1] = append(bindings[1], m.keys.Select, m.keys.NextField, m.keys.PrevField, m.keys.LeaveEditor)

	return bindings
}

func (m bookmarkEditorModel) setSize(width, height int) childModel {
	m.width = width
	m.height = height

	for _, input := range m.inputs {
		input.Width = width
	}

	return m
}

func (m bookmarkEditorModel) inputFocused() bool {
	return true
}

func (m bookmarkEditorModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *bookmarkEditorModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m *bookmarkEditorModel) saveBookmark() tea.Cmd {
	addressURL, err := url.Parse(m.inputs[0].Value())
	if err != nil {
		return errorCmd(err)
	}

	if len(addressURL.Scheme) <= 0 {
		return errorCmd(fmt.Errorf("Address requires scheme (http/https)"))
	}

	addressURL.Path = ""
	addressURL.RawQuery = ""
	addressURL.User = nil
	addressURL.Fragment = ""
	addressURL.Host = strings.TrimSuffix(addressURL.Host, "/")

	path := m.inputs[1].Value()
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	path = strings.TrimSuffix(path, "/")

	newBookmark := source.Bookmark{
		Backend:  "caddy",
		Address:  addressURL.String(),
		Path:     path,
		Username: m.inputs[2].Value(),
		Password: m.inputs[3].Value(),
	}

	if m.createNew {
		m.config.Bookmarks = append(m.config.Bookmarks, newBookmark)
	} else {
		m.config.Bookmarks[m.bookmarkIndex] = newBookmark
	}

	return saveBookmarkCmd
}

func (m *bookmarkEditorModel) updateInputStyles() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focusIndex {
			// Set focused state
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = focusedStyle
			m.inputs[i].TextStyle = focusedStyle
			continue
		}
		// Remove focused state
		m.inputs[i].Blur()
		m.inputs[i].PromptStyle = blurredStyle
		m.inputs[i].TextStyle = blurredStyle
	}

	return tea.Batch(cmds...)
}

func (m *bookmarkEditorModel) clearInputs() {
	for i := range m.inputs {
		m.inputs[i].Reset()
	}
}

func (m bookmarkEditorModel) Update(msg tea.Msg) (childModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		for _, input := range m.inputs {
			input.Width = msg.Width
		}

	case newBookmarkMsg:
		m.createNew = true
		m.focusIndex = 0
		m.clearInputs()
		cmds = append(cmds, m.updateInputStyles())

	case editBookmarkMsg:
		m.bookmarkIndex = msg.bookmarkIndex
		m.createNew = false
		m.clearInputs()

		bookmark := m.config.Bookmarks[m.bookmarkIndex]

		m.inputs[0].SetValue(bookmark.Address)
		m.inputs[1].SetValue(bookmark.Path)
		m.inputs[2].SetValue(bookmark.Username)
		m.inputs[3].SetValue(bookmark.Password)

		m.focusIndex = 0
		cmds = append(cmds, m.updateInputStyles())

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.LeaveEditor):
			cmds = append(cmds, changeViewCmd("bookmarkPicker"))

		case key.Matches(msg, m.keys.NextField):
			m.focusIndex++

			if m.focusIndex > m.inputCount {
				m.focusIndex = 0
			}

		case key.Matches(msg, m.keys.PrevField):
			m.focusIndex--

			if m.focusIndex < 0 {
				m.focusIndex = m.inputCount
			}

		case key.Matches(msg, m.keys.Select):
			switch m.focusIndex {
			case m.inputCount - 1:
				cmds = append(cmds, m.saveBookmark())

			case m.inputCount:
				cmds = append(cmds, changeViewCmd("bookmarkPicker"))

			default:
				m.focusIndex++

				if m.focusIndex > m.inputCount {
					m.focusIndex = 0
				}
			}
		}
		cmds = append(cmds, m.updateInputStyles())
	}

	cmd = m.updateInputs(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m bookmarkEditorModel) inputsView() string {
	var (
		sections []string
	)

	for _, input := range m.inputs {
		inputView := lipgloss.NewStyle().Padding(0, 0, 0, 2).Render(input.View())
		sections = append(sections, inputView)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m bookmarkEditorModel) titleView() string {
	var title string
	var view string

	if m.createNew {
		title = "New Bookmark"
	} else {
		title = m.config.Bookmarks[m.bookmarkIndex].Title()
	}

	view += titleStyle.Render(title)

	return titleBarStyle.Render(view)
}

func (m bookmarkEditorModel) View() string {
	var (
		sections    []string
		buttons     []string
		availHeight = m.height
	)

	titleView := m.titleView()
	sections = append(sections, titleView)
	availHeight -= lipgloss.Height(titleView)

	var submitView string
	if m.focusIndex == m.inputCount-1 {
		submitView = focusedButtonStyle.Render("Submit")
	} else {
		submitView = blurredButtonStyle.Render("Submit")
	}
	buttons = append(buttons, submitView)

	var cancelView string
	if m.focusIndex == m.inputCount {
		cancelView = focusedButtonStyle.Render("Cancel")
	} else {
		cancelView = blurredButtonStyle.Render("Cancel")
	}
	buttons = append(buttons, cancelView)

	buttonsView := lipgloss.NewStyle().MarginLeft(1).Render(lipgloss.JoinHorizontal(lipgloss.Center, buttons...))
	availHeight -= lipgloss.Height(buttonsView)

	inputsView := lipgloss.NewStyle().Height(availHeight).Render(m.inputsView())

	sections = append(sections, inputsView)

	sections = append(sections, buttonsView)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func newBookmarkEditor(config *cfg.Config) bookmarkEditorModel {
	keys := bookmarkEditorKeymap{
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select/next"),
		),

		NextField: key.NewBinding(
			key.WithKeys("down", "tab"),
			key.WithHelp("↓/tab", "next"),
		),

		PrevField: key.NewBinding(
			key.WithKeys("up", "shift+tab"),
			key.WithHelp("↑/shift+tab", "prev"),
		),

		LeaveEditor: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "leave editor"),
		),
	}

	m := bookmarkEditorModel{
		config:     config,
		inputs:     make([]textinput.Model, 4),
		focusIndex: 0,
		inputCount: 5,
		keys:       keys,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.NewModel()
		t.CursorStyle = cursorStyle
		t.CharLimit = 64

		switch i {
		case 0:
			t.Prompt = "Address: "
			t.Placeholder = "https://files.hostname.tld"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle

		case 1:
			t.Prompt = "Path: "
			t.Placeholder = "/home/media"

		case 2:
			t.Prompt = "Username: "
			t.Placeholder = "root"

		case 3:
			t.Prompt = "Password: "
			t.Placeholder = "toor"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '*'
		}

		m.inputs[i] = t
	}

	return m
}
