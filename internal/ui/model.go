// Package ui provides the ui for application
package ui

import (
	"bubblegit/internal/git"

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

type branchesMsg []git.BranchInfo
type errMsg struct{ err error }

type Model struct {
	dir string

	focus       int
	branchFocus int
	branches    []git.BranchInfo
	diff        viewport.Model
	err         error

	ready    bool
	quitting bool
}

func NewModel(dir string) Model {
	return Model{
		dir: dir,
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		branches, err := git.Branches(m.dir)
		if err != nil {
			return errMsg{err}
		}
		return branchesMsg(branches)
	}
}
