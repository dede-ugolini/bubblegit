// Package ui provides the ui for application
package ui

import (
	"time"

	"bubblegit/internal/git"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

const (
	focusStag = iota
	focusBranch
	focusLog
	focusStash
	focusDiff
	focusCount
)

type (
	filesMsg    []git.FileStatus
	branchesMsg []git.BranchInfo
	logMsg      []git.LogEntry
	stashesMsg  []git.StashEntry
	diffMsg     struct{ diff string }
	errMsg      struct{ err error }
	tickMsg     struct{}
)

const (
	commitFocusSummary = iota
	commitFocusMessage
)

type commitPopup struct {
	commitSummary textinput.Model
	commitMessage textarea.Model
	focus         int
	active        bool
	reword        bool
	rewordHash    string

	// squash fields are set when this popup was opened to confirm folding a
	// marked range of log commits into one, mirroring reword/rewordHash.
	squash           bool
	squashOldestHash string
	squashCount      int
}

type stashBranchPopup struct {
	input  textinput.Model
	active bool
}

// mergeMode identifies which merge strategy to apply.
type mergeMode int

const (
	mergeModeFF mergeMode = iota
	mergeModeMerge
	mergeModeSquash
)

func (m mergeMode) String() string {
	switch m {
	case mergeModeFF:
		return "fast-forward"
	case mergeModeMerge:
		return "merge commit"
	case mergeModeSquash:
		return "squash commit"
	default:
		return "unknown"
	}
}

// gitArg returns the mode string passed to git.Merge.
func (m mergeMode) gitArg() string {
	switch m {
	case mergeModeFF:
		return "ff"
	case mergeModeMerge:
		return "merge"
	case mergeModeSquash:
		return "squash"
	default:
		return "ff"
	}
}

var mergeModes = []mergeMode{mergeModeFF, mergeModeMerge, mergeModeSquash}

type mergePopup struct {
	active bool
	idx    int

	// branch is the branch being merged into the current branch.
	branch string
}

// inputAction identifies which git action inputPopup should dispatch on
// submit - one popup, several purposes, same as commitPopup.reword.
type inputAction int

const (
	inputActionNewBranch inputAction = iota
	inputActionRenameBranch
	inputActionPushStash
)

type inputPopup struct {
	input  textinput.Model
	active bool
	action inputAction
	title  string

	// renameFrom is the branch being renamed, captured when the popup is
	// opened. It's read only by inputActionRenameBranch.
	renameFrom string
}

const (
	tagFocusName = iota
	tagFocusMessage
)

type tagPopup struct {
	tagName    textinput.Model
	tagMessage textarea.Model
	focus      int
	active     bool

	// hash is the commit being tagged, captured when the popup opens.
	hash string
}

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

	// squashMarking is true while the user is marking a range of commits in
	// the log panel to squash together; squashAnchor is the log index where
	// marking started. The current range is always
	// [min(squashAnchor, idxLog), max(squashAnchor, idxLog)].
	squashMarking bool
	squashAnchor  int

	stashes           []git.StashEntry
	idxStash          int
	stashHeight       int
	stashWidth        int
	stashClearConfirm bool

	// dropConfirm is true while the user is confirming dropping the most
	// recent commit (m.log[0]); dropSubject is its subject for display.
	dropConfirm  bool
	dropSubject  string

	tagPopup    tagPopup
	commitPopup commitPopup

	stashBranchPopup stashBranchPopup

	inputPopup inputPopup

	mergePopup mergePopup

	focus int
	diff  viewport.Model
	err   error

	width  int
	height int

	panelFullScreen bool

	// useDelta toggles between delta-rendered diffs and plain git output.
	useDelta bool

	themeName string
	theme     Theme

	ready    bool
	quitting bool
}

func NewModel(dir string) Model {
	return Model{
		dir: dir,
		commitPopup: commitPopup{
			commitSummary: textinput.New(),
			commitMessage: textarea.New(),
		},
		tagPopup: tagPopup{
			tagName:    textinput.New(),
			tagMessage: textarea.New(),
		},
		stashBranchPopup: stashBranchPopup{
			input: textinput.New(),
		},
		inputPopup: inputPopup{
			input: textinput.New(),
		},
		themeName: themeOrder[0],
		theme:     themesByName[themeOrder[0]],
		useDelta:  true,
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
		func() tea.Msg {
			stashes, err := git.Stashes(m.dir)
			if err != nil {
				return errMsg{err}
			}
			return stashesMsg(stashes)
		},
	)
}
