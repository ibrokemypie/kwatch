package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ibrokemypie/kwatch/pkg/cfg"
)

type errorMsg struct {
	err error
}

func (e errorMsg) Error() string { return e.err.Error() }

func errorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return errorMsg{err}
	}
}

type clearErrorMsg struct{}

func clearErrorCmd() tea.Msg {
	return clearErrorMsg{}
}

type mainModel struct {
	pickerModel pickerModel
	config      *cfg.Config
	err         error
}

func (m mainModel) Init() tea.Cmd {
	return m.pickerModel.Init()
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case errorMsg:
		m.err = msg

	case clearErrorMsg:
		m.err = nil
	}

	m.pickerModel, cmd = m.pickerModel.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	var view string
	view += m.pickerModel.View()

	if m.err != nil {
		view += fmt.Sprintf("\nAn error occured: %v", m.err)
	}

	return view
}

func NewProgram(config *cfg.Config) *tea.Program {
	m := mainModel{
		config:      config,
		pickerModel: newPicker(config),
		err:         nil,
	}

	p := tea.NewProgram(m, tea.WithMouseCellMotion(), tea.WithAltScreen())
	return p
}
