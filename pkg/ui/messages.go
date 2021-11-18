package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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

type changeViewMsg struct {
	newViewName string
}

func changeViewCmd(newView string) tea.Cmd {
	return func() tea.Msg {
		return changeViewMsg{newView}
	}
}

type updateOpenBookmarkMsg struct {
	newOpenBookmark int
}

func updateOpenBookmarkCmd(newOpenBookmark int) tea.Cmd {
	return func() tea.Msg {
		return updateOpenBookmarkMsg{newOpenBookmark}
	}
}

type editBookmarkMsg struct {
	bookmarkIndex int
}

func editBookmarkCmd(index int) tea.Cmd {
	return func() tea.Msg {
		return editBookmarkMsg{bookmarkIndex: index}
	}
}

type newBookmarkMsg struct{}

func newBookmarkCmd() tea.Msg {
	return newBookmarkMsg{}
}

type saveBookmarkMsg struct{}

func saveBookmarkCmd() tea.Msg {
	return saveBookmarkMsg{}
}

type initialiseListMsg struct{}

func initialiseListCmd() tea.Msg {
	return initialiseListMsg{}
}

type endListUpdateMsg struct {
	itemList []list.Item
	newPath  []string
}

type endFileOpenMsg struct{}
