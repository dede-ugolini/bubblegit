// Package ui provides the ui for application
package ui

import (
	"bubblegit/internal/git"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

const (
	focusBranch = iota
	focusStag
	focusLog
	focusStash
	focusCount
)

type (
	filesMsg    []git.FileStatus
	branchesMsg []git.BranchInfo
	errMsg      struct{ err error }
)

type Model struct {
	dir string

	files []git.FileStatus

	focus       int
	branchFocus int
	branches    []git.BranchInfo
	diff        viewport.Model
	err         error

	input        textinput.Model
	focusInput   bool
	renameBranch string

	ready    bool
	quitting bool
}

func NewModel(dir string) Model {
	return Model{
		dir:   dir,
		input: textinput.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			files, err := git.Status(m.dir)
			if err != nil {
				return errMsg{err}
			}
			return filesMsg(files)
		},
		func() tea.Msg {
			branches, err := git.Branches(m.dir)
			if err != nil {
				return errMsg{err}
			}
			return branchesMsg(branches)
		},
	)
}
