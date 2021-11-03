package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ibrokemypie/kwatch/pkg/cfg"
)

type errorMsg struct {
	err error
}

func errorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return errorMsg{err}
	}
}

type mainModel struct {
	pickerModel pickerModel
	config      *cfg.Config
}

func (m mainModel) Init() tea.Cmd {
	return m.pickerModel.Init()
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// switch msg := msg.(type) {
	// }

	m.pickerModel, cmd = m.pickerModel.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	var view string
	view += m.pickerModel.View()

	return view
}

func NewProgram(config *cfg.Config, initialPath []string) *tea.Program {
	m := mainModel{
		config:      config,
		pickerModel: newPicker(config, initialPath),
	}

	p := tea.NewProgram(m, tea.WithMouseCellMotion())
	return p
}
