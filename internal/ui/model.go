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
	aheadMsg    map[string]bool
	tagsMsg     map[string][]string
	stashesMsg  []git.StashEntry
	diffMsg     struct{ diff string }
	errMsg      struct{ err error }
	tickMsg     struct{}

	// notRepoMsg reports that the working directory is not a git
	// repository; repoReadyMsg that it is (or has just become one) and the
	// panels should refresh.
	notRepoMsg   struct{}
	repoReadyMsg struct{}
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

// defaultConfirmHelp is the key hint shown when a confirm doesn't set
// its own.
const defaultConfirmHelp = "y/enter confirm · n/esc cancel"

// confirm is a modal "are you sure?" prompt for an action that discards
// work. Only one can be open at a time, so a single value on Model serves
// every such action; ask opens one and updateConfirm runs it.
//
// Whatever the prompt needs to display is captured in title and detail
// when it's opened, because the selection can move underneath it and the
// action must still refer to what the user was actually asked about.
type confirm struct {
	active bool

	// title is the question ("Delete branch 'foo'?"). detail is optional
	// extra context shown below it, blank for most prompts.
	title  string
	detail string

	// help is the key hint line. Empty means defaultConfirmHelp.
	help string

	// onYes runs the confirmed action, onNo the declined one. A nil action
	// just closes the prompt.
	//
	// These are tea.Cmds, so the git calls they run never block the UI.
	// Model is always used through a pointer, so the model a closure
	// captures is the live one and changes it makes are visible to the
	// next render.
	onYes tea.Cmd
	onNo  tea.Cmd
}

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
	inputActionSetRemote
	inputActionInitRepo
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

// panel is the geometry of one focusable list panel. The files, branch,
// log and stash panels show different data but are laid out identically,
// so their size lives here instead of a height/width pair per panel.
type panel struct {
	height, width int

	// idx is the selected row within the panel.
	idx int
}

// clampRow bounds idx to a list of n rows, matching the order the panels
// have always used: a below-range index clamps to 0 first, and one at or
// past the end clamps to n-1, so an empty list yields -1 when idx wasn't
// itself negative.
func clampRow(idx, n int) int {
	if idx < 0 {
		return 0
	}
	if idx >= n {
		return n - 1
	}
	return idx
}

type Model struct {
	dir string

	// panels holds the geometry and cursor of each list panel, indexed by
	// focus constant: panels[focusBranch] is the branch panel. focusDiff has
	// no entry, because the diff panel's size lives in the viewport itself
	// and it has no cursor.
	panels [focusCount]panel

	files []git.FileStatus

	// treeRows is the flattened file tree (directories interspersed with
	// files) the files panel renders; the files panel's cursor indexes it
	// rather than files. collapsed tracks which directory paths
	// (repo-relative) are collapsed, so directories are expanded by default.
	treeRows  []fileRow
	collapsed map[string]bool

	branches []git.BranchInfo

	log []git.LogEntry

	// ahead is the set of log hashes the current branch carries but its
	// upstream doesn't; their short hashes render in theme.Ahead.
	ahead map[string]bool

	tags map[string][]string

	// squashMarking is true while the user is marking a range of commits in
	// the log panel to squash together; squashAnchor is the log index where
	// marking started. The current range is always the span between
	// squashAnchor and the log panel's cursor.
	squashMarking bool
	squashAnchor  int

	stashes []git.StashEntry

	// confirm is the modal "are you sure?" prompt. At most one is open at
	// a time, so one value covers every destructive action instead of a
	// bool plus a captured detail string per action.
	confirm confirm

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

	// inRepo is false until a repository exists - while the startup
	// confirm is still up, or the branch name for git init hasn't been
	// entered yet. It gates the tick refresh to keep git errors off the
	// screen.
	inRepo bool
}

func NewModel(dir string) *Model {
	return &Model{
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
		collapsed: make(map[string]bool),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			if git.IsRepo(m.dir) {
				return repoReadyMsg{}
			}
			return notRepoMsg{}
		},
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// splitLayout is the share of the window each list panel gets in the split
// view, as a percentage of height and of width. The four heights total 97%,
// leaving room for the footer; the widths are all the same because the
// panels are stacked down the left of the window.
var splitLayout = [...]struct{ h, w int }{
	focusStag:   {25, 45},
	focusBranch: {15, 45},
	focusLog:    {42, 45},
	focusStash:  {15, 45},
}

// The diff panel's share of the window in the split view, as percentages
// of height and width.
const diffLayoutH, diffLayoutW = 90, 55

// setSplitLayout lays the window out as the four list panels stacked down
// the left with the diff alongside them. Resizing the window and leaving
// fullscreen both come through here, so the two can't drift apart.
func (m *Model) setSplitLayout(w, h int) {
	for i, f := range splitLayout {
		// Assign the geometry rather than the whole panel: a resize must
		// leave each panel's cursor where the user left it.
		p := &m.panels[i]
		p.height, p.width = h*f.h/100, w*f.w/100
	}
	m.diff.SetHeight(h * diffLayoutH / 100)
	m.diff.SetWidth(w * diffLayoutW / 100)
}

// setFullscreenLayout gives the whole window to the focused panel and
// collapses every other one to nothing, so the focused panel is the only
// thing left on screen.
func (m *Model) setFullscreenLayout(focus, w, h int) {
	for i := range m.panels {
		p := &m.panels[i]
		switch {
		case i == focusDiff:
			// The diff panel's size lives in the viewport, handled below.
		case i == focus:
			p.height, p.width = h, w
		default:
			p.height, p.width = 0, 0
		}
	}
	if focus == focusDiff {
		m.diff.SetHeight(h)
		m.diff.SetWidth(w)
		return
	}
	m.diff.SetHeight(0)
	m.diff.SetWidth(0)
}

// rowCount reports how many rows a list panel currently has. Each panel
// counts a different slice: the files panel counts the flattened tree, not
// m.files.
func (m *Model) rowCount(focus int) int {
	switch focus {
	case focusStag:
		return len(m.treeRows)
	case focusBranch:
		return len(m.branches)
	case focusLog:
		return len(m.log)
	case focusStash:
		return len(m.stashes)
	}
	return 0
}

// move moves a list panel's cursor by delta rows, clamped to the rows it
// has. focusDiff has no cursor and is left alone.
func (m *Model) move(focus, delta int) {
	if focus == focusDiff {
		return
	}
	m.panels[focus].idx = clampRow(m.panels[focus].idx+delta, m.rowCount(focus))
}

// clampCursors re-clamps every list panel's cursor. Call it after replacing
// a panel's data, so a list that shrank (or was empty when the cursor last
// moved) can't leave the cursor pointing past its end.
func (m *Model) clampCursors() {
	for i := focusStag; i < focusDiff; i++ {
		m.panels[i].idx = clampRow(m.panels[i].idx, m.rowCount(i))
	}
}

func (m *Model) Refresh() tea.Cmd {
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
			ahead, err := git.AheadHashes(m.dir)
			if err != nil {
				return errMsg{err}
			}
			return aheadMsg(ahead)
		},
		func() tea.Msg {
			tags, err := git.TagsByCommit(m.dir)
			if err != nil {
				return errMsg{err}
			}
			return tagsMsg(tags)
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
