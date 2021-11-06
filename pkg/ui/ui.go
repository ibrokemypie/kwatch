package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ibrokemypie/kwatch/pkg/cfg"
)

type viewName int

const (
	filePicker viewName = iota
	bookmarkPicker
	bookmarkEditor
)

type mainModel struct {
	config       *cfg.Config
	confFilePath string
	viewName
	filePickerModel
	bookmarkPickerModel
	bookmarkEditorModel
	err error
}

func (m mainModel) Init() tea.Cmd {
	return m.filePickerModel.Init()
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case errorMsg:
		m.err = msg

	case clearErrorMsg:
		m.err = nil

	case changeViewMsg:
		m.viewName = msg.newView
		cmds = append(cmds, clearErrorCmd)

	case newBookmarkMsg, editBookmarkMsg:
		m.viewName = bookmarkEditor

	case saveBookmarkMsg:
		err := m.config.WriteConfig(m.confFilePath)
		if err != nil {
			cmds = append(cmds, errorCmd(err))
		}

		m.viewName = bookmarkPicker

	case updateOpenBookmarkMsg:
		m.viewName = filePicker
	}

	switch m.viewName {
	case bookmarkPicker:
		m.bookmarkPickerModel, cmd = m.bookmarkPickerModel.Update(msg)

	case filePicker:
		m.filePickerModel, cmd = m.filePickerModel.Update(msg)

	case bookmarkEditor:
		m.bookmarkEditorModel, cmd = m.bookmarkEditorModel.Update(msg)
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	var view string

	switch m.viewName {
	case bookmarkPicker:
		view += m.bookmarkPickerModel.View()
	case filePicker:
		view += m.filePickerModel.View()
	case bookmarkEditor:
		view += m.bookmarkEditorModel.View()
	}

	if m.err != nil {
		view += fmt.Sprintf("\nAn error occured: %v", m.err)
	}

	return view
}

func NewProgram(config *cfg.Config, confFilePath string) *tea.Program {
	var currentView viewName
	if len(config.Bookmarks) > 0 {
		currentView = filePicker
	} else {
		currentView = bookmarkPicker
	}

	m := mainModel{
		config:              config,
		confFilePath:        confFilePath,
		viewName:            currentView,
		filePickerModel:     newFilePicker(config),
		bookmarkPickerModel: newBookmarkPicker(config),
		bookmarkEditorModel: newBookmarkEditor(config),
		err:                 nil,
	}

	p := tea.NewProgram(m, tea.WithMouseCellMotion(), tea.WithAltScreen())
	return p
}
