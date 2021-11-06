package ui

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ibrokemypie/kwatch/pkg/cfg"
	"github.com/ibrokemypie/kwatch/pkg/source"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle.Copy()
	noStyle      = lipgloss.NewStyle()
	helpStyle    = blurredStyle.Copy()
)


type bookmarkEditorModel struct {
	config        *cfg.Config
	bookmarkIndex int
	createNew     bool
	inputs        []textinput.Model
	inputCount    int
	focusIndex    int
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
		m.inputs[i].PromptStyle = noStyle
		m.inputs[i].TextStyle = noStyle
	}

	return tea.Batch(cmds...)
}

func (m bookmarkEditorModel) Update(msg tea.Msg) (bookmarkEditorModel, tea.Cmd) {
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
		cmds = append(cmds, m.updateInputStyles())

	case editBookmarkMsg:
		m.bookmarkIndex = msg.bookmarkIndex
		m.createNew = false

		bookmark := m.config.Bookmarks[m.bookmarkIndex]

		m.inputs[0].SetValue(bookmark.Address)
		m.inputs[1].SetValue(bookmark.Path)
		m.inputs[2].SetValue(bookmark.Username)
		m.inputs[3].SetValue(bookmark.Password)

		m.focusIndex = 0
		cmds = append(cmds, m.updateInputStyles())

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			cmds = append(cmds, changeViewCmd(bookmarkPicker))

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Pressed enter on submit button
			if s == "enter" && m.focusIndex == m.inputCount-1 {
				cmds = append(cmds, m.saveBookmark())
			}

			if s == "enter" && m.focusIndex == m.inputCount {
				cmds = append(cmds, changeViewCmd(bookmarkPicker))
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > m.inputCount {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = m.inputCount
			}

			cmds = append(cmds, m.updateInputStyles())
		}
	}

	cmd = m.updateInputs(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m bookmarkEditorModel) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	submitButton := blurredStyle.Render("Submit")
	if m.focusIndex == m.inputCount-1 {
		submitButton = focusedStyle.Render("Submit")
	}

	cancelButton := blurredStyle.Render("Cancel")
	if m.focusIndex == m.inputCount {
		cancelButton = focusedStyle.Render("Cancel")
	}
	fmt.Fprintf(&b, "\n\n%s\t%s\n\n", submitButton, cancelButton)

	return b.String()
}

func newBookmarkEditor(config *cfg.Config) bookmarkEditorModel {
	m := bookmarkEditorModel{
		config:     config,
		inputs:     make([]textinput.Model, 4),
		focusIndex: 0,
		inputCount: 5,
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
