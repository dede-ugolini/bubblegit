// Package ui provides the ui for application
package ui

import (
	"time"

	"bubblegit/internal/git"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

const (
	focusStag = iota
	focusBranch
	focusLog
	focusDiff
	focusCount
)

type (
	filesMsg    []git.FileStatus
	branchesMsg []git.BranchInfo
	logMsg      []git.LogEntry
	diffMsg     struct{ diff string }
	errMsg      struct{ err error }
	tickMsg     struct{}
)

type Model struct {
	dir string

	files       []git.FileStatus
	idxFiles    int
	filesHeight int
	filesWidth  int

	branches     []git.BranchInfo
	idxBranch    int
	branchHeight int
	branchWidth  int

	log       []git.LogEntry
	idxLog    int
	logHeight int
	logWidth  int

	focus int
	diff  viewport.Model
	err   error

	width  int
	height int

	panelFullScreen bool

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
	return tea.Batch(m.Refresh(), tickCmd())
}

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m Model) Refresh() tea.Cmd {
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
		func() tea.Msg {
			log, err := git.Log(m.dir, "HEAD", 200)
			if err != nil {
				return errMsg{err}
			}
			return logMsg(log)
		},
	)
}
