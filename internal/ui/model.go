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

	// onYes runs the confirmed action, onNo the declined one. Both are
	// called from Update with the model that is about to be returned, so
	// an action can open another popup; the Cmd it hands back is the work
	// itself, which is where the git calls live so they never block the
	// UI. A nil action just closes the prompt.
	//
	// The action is run during Update rather than returned as a Cmd,
	// because Update has a value receiver: a model captured in a Cmd
	// outlives the call that would have made its mutations visible.
	onYes func(*Model) tea.Cmd
	onNo  func(*Model) tea.Cmd
}

// deferred adapts a handler for use as a confirm action. The handler only
// reads the model, so it can run as a Cmd against the snapshot taken when
// the prompt opened - the same selection the question refers to.
func deferred(h tea.Cmd) func(*Model) tea.Cmd {
	return func(*Model) tea.Cmd { return h }
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

type Model struct {
	dir string

	files       []git.FileStatus
	idxFiles    int
	filesHeight int
	filesWidth  int

	// treeRows is the flattened file tree (directories interspersed with
	// files) the files panel renders; idxFiles indexes it rather than files.
	// collapsed tracks which directory paths (repo-relative) are collapsed,
	// so directories are expanded by default.
	treeRows  []fileRow
	collapsed map[string]bool

	branches     []git.BranchInfo
	idxBranch    int
	branchHeight int
	branchWidth  int

	log       []git.LogEntry
	idxLog    int
	logHeight int
	logWidth  int

	// ahead is the set of log hashes the current branch carries but its
	// upstream doesn't; their short hashes render in theme.Ahead.
	ahead map[string]bool

	tags map[string][]string

	// squashMarking is true while the user is marking a range of commits in
	// the log panel to squash together; squashAnchor is the log index where
	// marking started. The current range is always
	// [min(squashAnchor, idxLog), max(squashAnchor, idxLog)].
	squashMarking bool
	squashAnchor  int

	stashes     []git.StashEntry
	idxStash    int
	stashHeight int
	stashWidth  int

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
		collapsed: make(map[string]bool),
	}
}

func (m Model) Init() tea.Cmd {
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
