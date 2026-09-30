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

	// detachedMsg carries the short hash HEAD points at when it is not on a
	// local branch, and the empty string when it is. Since it is sent on
	// every refresh rather than only on the transition, the model always
	// clears a stale hash when HEAD is reattached to a branch.
	detachedMsg string

	// notRepoMsg reports that the working directory is not a git
	// repository; repoReadyMsg that it is (or has just become one) and the
	// panels should refresh.
	notRepoMsg   struct{}
	repoReadyMsg struct{}

	// conflictMsg carries a conflicted file's hunks back from the git layer,
	// or the reason they could not be read. Reading them runs git and touches
	// temp files, so it cannot happen inline on a key press.
	conflictMsg struct {
		path string
		file *git.ConflictFile
		err  error
	}
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

// conflictPopup resolves a conflicted file one hunk at a time.
//
// It holds the whole file's conflicts rather than the one under the cursor,
// because picking a side is per hunk but saving is per file: there is no
// partial resolution to commit, so the hunk list and the detail of the
// selected hunk are two views of one decision.
type conflictPopup struct {
	active bool

	// path is the file being resolved and file its parsed hunks, both
	// captured when the popup was opened. The files panel selection can move
	// underneath, and the resolve has to stay about the file it was opened
	// for.
	path string
	file *git.ConflictFile

	// choices is one resolution per hunk. Every hunk starts on ResolveOurs
	// rather than unset, so that saving without touching anything keeps our
	// side of the merge - which is what entering and pressing enter again
	// should mean - instead of failing at the last step.
	choices []git.Resolution

	// idx is the hunk the cursor is on, and scroll the first line of that
	// hunk's detail pane that is visible.
	idx    int
	scroll int
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

	// detached is the short hash HEAD points at when it is not on a local
	// branch, and the empty string when it is. While non-empty no branch in
	// the branch panel is marked current, so the footer has to say where the
	// working tree actually is.
	detached string

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

	conflictPopup conflictPopup

	focus int
	diff  viewport.Model
	err   error

	width  int
	height int

	// mode is how much of the window the focused panel gets: the whole
	// window in normal mode, half of it in middle, all of it in fullscreen.
	mode layoutMode

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

// layoutMode is how much of the window the focused panel gets. The modes are
// ordered by how much room they give it, so that stepping with + and - walks
// the sequence in both directions.
type layoutMode int

const (
	// layoutNormal stacks the four list panels down the left with the diff
	// beside them, and is where the app starts.
	layoutNormal layoutMode = iota
	// layoutMiddle gives the focused panel the left half of the window at full
	// height, with the diff in the right half.
	layoutMiddle
	// layoutFull gives the focused panel the whole window.
	layoutFull
)

// step moves n modes along the sequence, reporting false if that would run off
// either end, so + and - stop at the extremes rather than wrapping.
func (m layoutMode) step(n int) (layoutMode, bool) {
	next := int(m) + n
	if next < 0 || next > int(layoutFull) {
		return m, false
	}
	return layoutMode(next), true
}

// full reports whether the focused panel has the whole window to itself. This
// is what tells the diff functions to render side by side: there's only room
// for that when the diff has the window, not half of it.
func (m layoutMode) full() bool { return m == layoutFull }

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

// setMiddleLayout gives the focused panel the left half of the window at full
// height and the diff the right half, collapsing the three list panels that
// aren't focused. A focused diff takes the left half instead and leaves the
// right half empty, so the focused panel is always the one on the left.
func (m *Model) setMiddleLayout(focus, w, h int) {
	left := w / 2
	for i := focusStag; i < focusDiff; i++ {
		p := &m.panels[i]
		// Geometry only, so a mode change leaves each cursor where the user
		// left it - same as the other two layouts.
		if i == focus {
			// A list panel's width is its body's: the border adds the two
			// columns on either side, so the box needs the half-width less
			// them to come out at half the window. Clamped at 1, since a
			// zero or negative width renders the panel as nothing at all -
			// better a too-narrow box than no box.
			p.height, p.width = h, max(left-2, 1)
			continue
		}
		p.height, p.width = 0, 0
	}
	// The diff panel is the one panel whose width is the whole box, since its
	// viewport is padded to it. Its height, by contrast, is the viewport
	// inside the box, so the two border rows have to come off the window
	// height to leave a box of h.
	m.diff.SetHeight(max(h-2, 1))
	if focus == focusDiff {
		// The diff is the focused panel, so it's the one on the left.
		m.diff.SetWidth(max(left, 1))
		return
	}
	m.diff.SetWidth(max(w-left, 1))
}

// setLayout applies mode to the current window size. Resizing and changing
// mode both come through here, so the three layouts can't drift apart - and
// so a resize keeps whatever mode the user was in, which a direct
// setSplitLayout call on WindowSizeMsg would not.
func (m *Model) setLayout(mode layoutMode) {
	m.mode = mode
	switch mode {
	case layoutMiddle:
		m.setMiddleLayout(m.focus, m.width, m.height)
	case layoutFull:
		m.setFullscreenLayout(m.focus, m.width, m.height)
	default:
		m.setSplitLayout(m.width, m.height)
	}
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
		func() tea.Msg {
			// Detection never fails, so this never reports an error: an
			// attached HEAD just yields the empty string.
			if hash, ok := git.Detached(m.dir); ok {
				return detachedMsg(hash)
			}
			return detachedMsg("")
		},
	)
}
