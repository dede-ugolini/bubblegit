package ui

import (
	"fmt"
	"strings"

	"bubblegit/internal/git"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// Update handles a message and returns the updated model.
//
// Model is used through a pointer throughout, and that is load-bearing:
// commands run after Update returns, so a model captured in a tea.Cmd is
// only the live one if every method shares the same pointer. With a value
// receiver a closure would mutate a copy that is already gone by the time
// it runs, and the change would be silently dropped.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.squashMarking {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "up", "k":
				m.move(focusLog, -1)
				return m, m.showDiff()
			case "down", "j":
				m.move(focusLog, 1)
				return m, m.showDiff()
			case "S":
				lo, hi := m.squashAnchor, m.panels[focusLog].idx
				if lo > hi {
					lo, hi = hi, lo
				}
				count := hi - lo + 1
				m.squashMarking = false
				if count < 2 {
					return m, func() tea.Msg { return errMsg{fmt.Errorf("select at least two commits to squash")} }
				}
				return m, m.startSquash(lo, hi)
			case "esc":
				m.squashMarking = false
				return m, nil
			}
		}
		return m, nil
	}

	if m.confirm.active {
		return m, m.updateConfirm(msg)
	}

	if m.mergePopup.active {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "up", "k":
				m.mergePopup.idx = (m.mergePopup.idx - 1 + len(git.MergeModes)) % len(git.MergeModes)
				return m, nil
			case "down", "j":
				m.mergePopup.idx = (m.mergePopup.idx + 1) % len(git.MergeModes)
				return m, nil
			case "enter":
				mode := git.MergeModes[m.mergePopup.idx]
				branch := m.mergePopup.branch
				m.mergePopup.active = false
				return m, tea.Sequence(func() tea.Msg { return m.handleMerge(branch, mode) }, m.Refresh())
			case "esc":
				m.mergePopup.active = false
				return m, nil
			}
		}
		return m, nil
	}

	if m.stashBranchPopup.active {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter":
				name := m.stashBranchPopup.input.Value()
				if name == "" {
					return m, nil
				}
				m.stashBranchPopup.input.Blur()
				m.stashBranchPopup.active = false
				return m, tea.Sequence(m.handleStashBranch, m.Refresh())
			case "esc":
				m.stashBranchPopup.input.Blur()
				m.stashBranchPopup.active = false
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.stashBranchPopup.input, cmd = m.stashBranchPopup.input.Update(msg)
		return m, cmd
	}

	if m.commitPopup.active {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "ctrl+s":
				if m.commitPopup.commitSummary.Value() == "" {
					return m, nil
				}
				m.commitPopup.commitSummary.Blur()
				m.commitPopup.commitMessage.Blur()
				m.commitPopup.active = false
				return m, tea.Sequence(m.handleCommit, m.Refresh())
			case "tab":
				if m.commitPopup.focus == commitFocusSummary {
					m.commitPopup.commitSummary.Blur()
					m.commitPopup.commitMessage.Focus()
					m.commitPopup.focus = commitFocusMessage
				} else {
					m.commitPopup.commitMessage.Blur()
					m.commitPopup.commitSummary.Focus()
					m.commitPopup.focus = commitFocusSummary
				}
				return m, nil
			case "esc":
				m.commitPopup.commitSummary.Blur()
				m.commitPopup.commitMessage.Blur()
				m.commitPopup.active = false
				return m, nil
			case "enter":
				if m.commitPopup.focus == commitFocusSummary {
					m.commitPopup.commitSummary.Blur()
					m.commitPopup.commitMessage.Focus()
					m.commitPopup.focus = commitFocusMessage
				}
				return m, nil
			}
		}
		var cmd tea.Cmd
		if m.commitPopup.focus == commitFocusSummary {
			m.commitPopup.commitSummary, cmd = m.commitPopup.commitSummary.Update(msg)
			return m, cmd
		}
		m.commitPopup.commitMessage, cmd = m.commitPopup.commitMessage.Update(msg)
		return m, cmd
	}

	if m.tagPopup.active {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "ctrl+s":
				if m.tagPopup.tagName.Value() == "" {
					return m, nil
				}
				m.tagPopup.tagName.Blur()
				m.tagPopup.tagMessage.Blur()
				m.tagPopup.active = false
				return m, tea.Sequence(m.handleTag, m.Refresh())
			case "tab":
				if m.tagPopup.focus == tagFocusName {
					m.tagPopup.tagName.Blur()
					m.tagPopup.tagMessage.Focus()
					m.tagPopup.focus = tagFocusMessage
				} else {
					m.tagPopup.tagMessage.Blur()
					m.tagPopup.tagName.Focus()
					m.tagPopup.focus = tagFocusName
				}
				return m, nil
			case "esc":
				m.tagPopup.tagName.Blur()
				m.tagPopup.tagMessage.Blur()
				m.tagPopup.active = false
				return m, nil
			case "enter":
				if m.tagPopup.focus == tagFocusName {
					m.tagPopup.tagName.Blur()
					m.tagPopup.tagMessage.Focus()
					m.tagPopup.focus = tagFocusMessage
				}
				return m, nil
			}
		}
		var cmd tea.Cmd
		if m.tagPopup.focus == tagFocusName {
			m.tagPopup.tagName, cmd = m.tagPopup.tagName.Update(msg)
			return m, cmd
		}
		m.tagPopup.tagMessage, cmd = m.tagPopup.tagMessage.Update(msg)
		return m, cmd
	}

	if m.inputPopup.active {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter":
				value := m.inputPopup.input.Value()
				action := m.inputPopup.action
				renameFrom := m.inputPopup.renameFrom
				m.inputPopup.input.Blur()
				m.inputPopup.active = false
				switch action {
				case inputActionPushStash:
					// Message is optional, so an empty value still submits.
					return m, m.handlePushStash
				case inputActionRenameBranch:
					if value == "" {
						return m, nil
					}
					return m, func() tea.Msg { return m.handleRenameBranch(renameFrom) }
				case inputActionNewBranch:
					if value == "" {
						return m, nil
					}
					return m, m.handleCreateBranch
				case inputActionSetRemote:
					if value == "" {
						return m, nil
					}
					return m, m.handleSetRemote
				case inputActionInitRepo:
					// Empty branch name falls back to git's default.
					return m, func() tea.Msg { return m.handleInitRepo(value) }
				}
				return m, nil
			case "esc":
				if m.inputPopup.action == inputActionInitRepo {
					return m, tea.Quit
				}
				m.inputPopup.input.Blur()
				m.inputPopup.active = false
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.inputPopup.input, cmd = m.inputPopup.input.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {

	case errMsg:
		m.err = msg.err
		return m, nil

	case notRepoMsg:
		m.inRepo = false
		m.ask(confirm{
			title: "Not a git repository. Create a new git repository?",
			help:  "y/enter create · n/esc quit",
			// Confirming doesn't init directly - it asks for the branch
			// name first, so this opens the input popup rather than
			// running a git command.
			onYes: func() tea.Msg {
				m.inputPopup.action = inputActionInitRepo
				m.inputPopup.title = "Branch name? (leave empty for git's default)"
				m.inputPopup.input.SetWidth(m.width / 3)
				m.inputPopup.input.CharLimit = 100
				m.inputPopup.input.SetValue("")
				m.inputPopup.input.Prompt = "branch> "
				m.inputPopup.input.Placeholder = "branch name (optional)"
				m.inputPopup.input.Focus()
				m.inputPopup.active = true
				return textinput.Blink
			},
			// Declining leaves nothing to show, so quit rather than
			// sitting on an empty window.
			onNo: tea.Quit,
		})
		return m, nil

	case repoReadyMsg:
		m.inRepo = true
		return m, m.Refresh()

	case tickMsg:
		// While no repository exists yet (still asking to create one or
		// entering the branch name) the git commands would all fail, so
		// keep refreshing on the timer alone.
		if !m.inRepo {
			return m, tickCmd()
		}
		return m, tea.Sequence(m.Refresh(), tickCmd())

	case filesMsg:
		m.files = []git.FileStatus(msg)
		m.rebuildFileTree()
		return m, nil

	case branchesMsg:
		m.branches = []git.BranchInfo(msg)
		m.clampCursors()
		return m, nil

	case logMsg:
		m.log = []git.LogEntry(msg)
		m.clampCursors()
		return m, nil

	case aheadMsg:
		m.ahead = map[string]bool(msg)
		return m, nil

	case tagsMsg:
		m.tags = map[string][]string(msg)
		return m, nil

	case stashesMsg:
		m.stashes = []git.StashEntry(msg)
		m.clampCursors()
		return m, nil

	case detachedMsg:
		m.detached = string(msg)
		return m, nil

	case diffMsg:
		m.diff.SetContent(msg.diff)
		return m, nil

	case tea.MouseClickMsg:
		m.mouseClick(msg.Mouse())
		return m, m.showDiff()

	case tea.MouseWheelMsg:
		return m, m.mouseWheel(msg.Mouse())

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.setLayout(m.mode)
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		m.err = nil
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			m.focus = (m.focus + 1) % focusCount
			return m, m.showDiff()

		case "0":
			m.focus = focusDiff

		case "1":
			m.focus = focusStag
			return m, m.showDiff()

		case "2":
			m.focus = focusBranch
			return m, m.showDiff()

		case "3":
			m.focus = focusLog
			return m, m.showDiff()

		case "4":
			m.focus = focusStash
			return m, m.showDiff()

		case "up", "k":
			if m.focus == focusDiff {
				var cmd tea.Cmd
				m.diff, cmd = m.diff.Update(msg)
				return m, cmd
			}
			m.move(m.focus, -1)
			return m, m.showDiff()

		case "down", "j":
			if m.focus == focusDiff {
				var cmd tea.Cmd
				m.diff, cmd = m.diff.Update(msg)
				return m, cmd
			}
			m.move(m.focus, 1)
			return m, m.showDiff()

		case "d":
			// Delete Branch
			if m.focus == focusBranch && len(m.branches) > 0 {
				m.ask(confirm{
					title: "Delete branch '" + m.branches[m.panels[focusBranch].idx].Name + "'?",
					onYes: m.handleDeleteBranch,
				})
			}
			// Restore file
			if m.focus == focusStag {
				if idx, ok := m.selectedFile(); ok {
					m.ask(confirm{
						title: "Restore '" + m.files[idx].Path + "'?",
						onYes: m.handleRestoreFile,
					})
				}
			}
			// Drop stash
			if m.focus == focusStash && len(m.stashes) > 0 {
				m.ask(confirm{
					title: "Drop stash '" + m.stashes[m.panels[focusStash].idx].Message + "'?",
					onYes: m.handleDropStash,
				})
			}
			// Drop last commit
			if m.focus == focusLog && len(m.log) > 0 {
				m.ask(confirm{
					title:  "Drop last commit?",
					detail: m.log[0].Subject,
					onYes:  m.handleDropCommit,
				})
			}

		case "D":
			// Clear Stash
			if m.focus == focusStash && len(m.stashes) > 0 {
				m.ask(confirm{
					title: "Remove all stash entries?",
					onYes: m.handleClearStash,
				})
			}

		case "b":
			// Stash branch
			if m.focus == focusStash && len(m.stashes) > 0 {
				m.stashBranchPopup.input.SetValue("")
				m.stashBranchPopup.input.SetWidth(m.width / 3)
				m.stashBranchPopup.input.CharLimit = 100
				m.stashBranchPopup.input.Placeholder = "branch name"
				m.stashBranchPopup.input.Focus()
				m.stashBranchPopup.active = true
				return m, textinput.Blink
			}

		case "n":
			// Create Branch
			if m.focus == focusBranch {
				m.inputPopup.action = inputActionNewBranch
				m.inputPopup.title = "New branch"
				m.inputPopup.input.SetWidth(m.width / 3)
				m.inputPopup.input.CharLimit = 20
				m.inputPopup.input.SetValue("")
				m.inputPopup.input.Prompt = "branch> "
				m.inputPopup.input.Placeholder = "new branch name"
				m.inputPopup.input.Focus()
				m.inputPopup.active = true
				return m, textinput.Blink
			}
			// Push stash
			if m.focus == focusStash {
				m.inputPopup.action = inputActionPushStash
				m.inputPopup.title = "New stash"
				m.inputPopup.input.SetWidth(m.width / 3)
				m.inputPopup.input.CharLimit = 72
				m.inputPopup.input.SetValue("")
				m.inputPopup.input.Prompt = "stash> "
				m.inputPopup.input.Placeholder = "stash message (optional)"
				m.inputPopup.input.Focus()
				m.inputPopup.active = true
				return m, textinput.Blink
			}

		case "r":
			// Rename Branch
			if m.focus == focusBranch && len(m.branches) > 0 {
				old := m.branches[m.panels[focusBranch].idx].Name
				m.inputPopup.action = inputActionRenameBranch
				m.inputPopup.title = "Rename branch"
				m.inputPopup.renameFrom = old
				m.inputPopup.input.SetWidth(m.width / 3)
				m.inputPopup.input.CharLimit = 20
				m.inputPopup.input.SetValue(old)
				m.inputPopup.input.Prompt = "rename> "
				m.inputPopup.input.Placeholder = "new name"
				m.inputPopup.input.Focus()
				m.inputPopup.active = true
				return m, textinput.Blink
			}
			// Reword commit
			if m.focus == focusLog && len(m.log) > 0 {
				return m, m.handleRewordCommit()
			}

		case "enter":
			// Expand/collapse file tree directory
			if m.focus == focusStag && len(m.treeRows) > 0 {
				if row := m.treeRows[m.panels[focusStag].idx]; row.isDir {
					if m.collapsed[row.dir] {
						delete(m.collapsed, row.dir)
					} else {
						m.collapsed[row.dir] = true
					}
					m.rebuildFileTree()
				}
			}
			// Checkout Branch
			if m.focus == focusBranch && len(m.branches) > 0 {
				return m, tea.Sequence(m.handleCheckoutBranch, m.Refresh())
			}
			// Apply Stash
			if m.focus == focusStash && len(m.stashes) > 0 {
				return m, m.handleApplyStash
			}
			// Checkout commit, leaving HEAD detached at it. Unlike the
			// destructive actions this asks for no confirmation, matching
			// enter-to-checkout a branch: git itself refuses the checkout if
			// it would discard local modifications.
			if m.focus == focusLog && len(m.log) > 0 {
				return m, tea.Sequence(m.handleCheckoutCommit, m.Refresh())
			}
		case "space":
			// stage/unstage file
			if m.focus == focusStag {
				if _, ok := m.selectedFile(); ok {
					return m, m.handleToggleStage
				}
			}
		case "a":
			// stage/unstage all
			if m.focus == focusStag && len(m.files) > 0 {
				return m, m.handleToggleStageAll
			}
		case "A":
			// Amend commit
			if m.focus == focusStag && len(m.files) > 0 && git.HasOneStaged(m.files) && len(m.log) > 0 {
				m.ask(confirm{
					title: "Amend commit '" + m.log[0].Subject + "'?",
					onYes: m.handleAmend,
				})
			}
		case "p":
			// Pop stash
			if m.focus == focusStash && len(m.stashes) > 0 {
				m.ask(confirm{
					title: "Pop stash '" + m.stashes[m.panels[focusStash].idx].Message + "'?",
					onYes: m.handlePopStash,
				})
			}
		case "c":
			// Commit
			if m.focus == focusStag && len(m.files) > 0 && git.HasOneStaged(m.files) {
				m.commitPopup.reword = false
				m.commitPopup.rewordHash = ""
				m.commitPopup.squash = false
				m.commitPopup.squashOldestHash = ""
				m.commitPopup.squashCount = 0
				m.commitPopup.focus = commitFocusSummary
				m.commitPopup.commitSummary.SetWidth(m.width / 3)
				m.commitPopup.commitMessage.SetWidth(m.width / 3)
				m.commitPopup.commitSummary.SetValue("")
				m.commitPopup.commitMessage.SetValue("")
				m.commitPopup.commitSummary.Placeholder = "Commit summary"
				m.commitPopup.commitMessage.Placeholder = "Commit message"
				m.commitPopup.commitSummary.Focus()
				m.commitPopup.commitMessage.Blur()
				m.commitPopup.active = true
				return m, textinput.Blink
			}

		// + and - step through the layouts: normal, middle, fullscreen. The
		// diff has to be re-rendered either way, since its width changes.
		case "+", "-":
			dir := 1
			if msg.String() == "-" {
				dir = -1
			}
			if next, ok := m.mode.step(dir); ok {
				m.setLayout(next)
				return m, m.showDiff()
			}

		case "pgup":
			m.diff.ScrollUp(8)
		case "pgdown":
			m.diff.ScrollDown(8)

		case "t":
			// Cycle theme
			m.themeName, m.theme = nextTheme(m.themeName)

		case "V":
			// Toggle diff rendering mode (delta vs plain git)
			m.useDelta = !m.useDelta
			return m, m.showDiff()

		case "S":
			// Squash: start marking a range of commits to fold together.
			if m.focus == focusLog && len(m.log) > 0 {
				m.squashMarking = true
				m.squashAnchor = m.panels[focusLog].idx
			}

		case "M":
			if m.focus == focusBranch && len(m.branches) > 0 {
				m.mergePopup.active = true
				m.mergePopup.idx = 0
				m.mergePopup.branch = m.branches[m.panels[focusBranch].idx].Name
			}

		case "P":
			if m.focus == focusBranch && len(m.branches) > 0 {
				branch := m.branches[m.panels[focusBranch].idx].Name
				return m, tea.Sequence(func() tea.Msg { return m.handlePush(branch) }, m.Refresh())
			}

		case "R":
			// Set the "origin" remote (adds it if the repo has none, or
			// repoints it if it does).
			if m.focus == focusBranch {
				m.inputPopup.action = inputActionSetRemote
				m.inputPopup.title = "Set remote"
				m.inputPopup.input.SetWidth(m.width / 3)
				m.inputPopup.input.CharLimit = 200
				m.inputPopup.input.SetValue("")
				m.inputPopup.input.Prompt = "remote> "
				m.inputPopup.input.Placeholder = "origin URL"
				m.inputPopup.input.Focus()
				m.inputPopup.active = true
				return m, textinput.Blink
			}

		case "T":
			if m.focus == focusLog && len(m.log) > 0 {
				m.tagPopup.active = true
				m.tagPopup.hash = m.log[m.panels[focusLog].idx].Hash
				m.tagPopup.focus = tagFocusName
				m.tagPopup.tagName.SetWidth(m.width / 3)
				m.tagPopup.tagMessage.SetWidth(m.width / 3)
				m.tagPopup.tagName.SetValue("")
				m.tagPopup.tagMessage.SetValue("")
				m.tagPopup.tagName.Placeholder = "Tag name"
				m.tagPopup.tagMessage.Placeholder = "Tag message (optional)"
				m.tagPopup.tagName.Focus()
				m.tagPopup.tagMessage.Blur()
				return m, textinput.Blink
			}
		}
	}

	return m, nil
}

func (m *Model) showDiff() tea.Cmd {
	return func() tea.Msg {
		switch m.focus {
		case focusStag:
			idx, ok := m.selectedFile()
			if !ok {
				return diffMsg{}
			}
			var (
				diff string
				err  error
			)
			file := m.files[idx]
			// Untracked() is also true for Unstaged() (an untracked file has
			// no staged changes, so its worktree side is by definition
			// unstaged) and a partially-staged file satisfies both Staged()
			// and Unstaged() - these are deliberately if/else-if, in
			// priority order, so exactly one diff is computed per file.
			if file.Staged() {
				if m.useDelta {
					diff, err = git.DiffDeltaStaged(m.dir, file.Path, m.mode.full(), m.diff.Width())
				} else {
					diff, err = git.DiffStaged(m.dir, file.Path)
				}
			} else if file.Untracked() {
				if m.useDelta {
					diff, err = git.DiffDeltaUntracked(m.dir, file.Path, m.mode.full(), m.diff.Width())
				} else {
					diff, err = git.DiffUntracked(m.dir, file.Path)
				}
			} else if file.Unstaged() {
				if m.useDelta {
					diff, err = git.DiffDelta(m.dir, file.Path, m.mode.full(), m.diff.Width())
				} else {
					diff, err = git.Diff(m.dir, file.Path)
				}
			}

			if err != nil {
				return errMsg{err}
			}
			return diffMsg{diff}
		case focusBranch:
			if len(m.branches) <= 0 {
				return diffMsg{}
			}
			var (
				diff string
				err  error
			)
			if m.useDelta {
				diff, err = git.DiffBranchDelta(m.dir, m.mode.full(), m.diff.Width())
			} else {
				diff, err = git.DiffBranch(m.dir)
			}
			if err != nil {
				return errMsg{err}
			}
			return diffMsg{diff}
		case focusLog:
			if len(m.log) == 0 {
				return diffMsg{}
			}
			var (
				diff string
				err  error
			)
			hash := m.log[m.panels[focusLog].idx].Hash
			if m.useDelta {
				diff, err = git.ShowDelta(m.dir, hash, m.mode.full(), m.diff.Width())
			} else {
				diff, err = git.Show(m.dir, hash)
			}
			if err != nil {
				return errMsg{err}
			}
			return diffMsg{diff}
		case focusStash:
			if len(m.stashes) == 0 {
				return diffMsg{}
			}
			var (
				diff string
				err  error
			)
			ref := m.stashes[m.panels[focusStash].idx].Ref
			if m.useDelta {
				diff, err = git.StashShowDelta(m.dir, ref, m.mode.full(), m.diff.Width())
			} else {
				diff, err = git.StashShow(m.dir, ref)
			}
			if err != nil {
				return errMsg{err}
			}
			return diffMsg{diff}
		}
		return nil
	}
}

// ask opens a confirmation prompt, replacing whatever was open (nothing
// else can be, since a prompt swallows every message while it's up).
func (m *Model) ask(c confirm) {
	c.active = true
	m.confirm = c
}

// updateConfirm handles a key while a confirmation prompt is open. It
// swallows every other message as well, so nothing can slip past the
// modal to the main switch behind it.
func (m *Model) updateConfirm(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	var run tea.Cmd
	switch key.String() {
	case "y", "enter":
		run = m.confirm.onYes
	case "n", "esc":
		run = m.confirm.onNo
	default:
		return nil
	}
	// Clear the whole struct, not just active: onYes holds the selection
	// the question refers to, and leaving it around would let a later
	// keypress fire it against a stale question.
	m.confirm = confirm{}
	return run
}

func (m *Model) handleDeleteBranch() tea.Msg {
	branch := m.branches[m.panels[focusBranch].idx].Name
	err := git.DeleteBranch(m.dir, branch)
	if err != nil {
		return errMsg{err}
	}
	branches, err := git.Branches(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return branchesMsg(branches)
}

func (m *Model) handleRestoreFile() tea.Msg {
	idx, ok := m.selectedFile()
	if !ok {
		return nil
	}
	if m.files[idx].Untracked() {
		err := git.RestoreUntracked(m.dir, m.files[idx].Path)
		if err != nil {
			return errMsg{err}
		}
	} else {
		err := git.Restore(m.dir, m.files[idx].Path)
		if err != nil {
			return errMsg{err}
		}
	}
	files, err := git.Status(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return filesMsg(files)
}

// handleCheckoutBranch checks out the branch panel's selected branch. As
// with handleCheckoutCommit the caller sequences a Refresh on after, since a
// checkout moves HEAD, the working tree and the index and so invalidates
// every panel, not just the branch list.
func (m *Model) handleCheckoutBranch() tea.Msg {
	if err := git.Checkout(m.dir, m.branches[m.panels[focusBranch].idx].Name); err != nil {
		return errMsg{err}
	}
	return nil
}

// handleCheckoutCommit checks the working tree out at the log panel's
// selected commit, leaving HEAD detached at it. The caller sequences a
// Refresh on after, so unlike the single-panel handlers this one does not
// re-read the branches it just invalidated.
//
// It guards on an empty log itself rather than relying on the key handler's
// len(m.log) > 0 check, so a stale keypress cannot index into nothing.
func (m *Model) handleCheckoutCommit() tea.Msg {
	if len(m.log) == 0 {
		return nil
	}
	if err := git.Checkout(m.dir, m.log[m.panels[focusLog].idx].Hash); err != nil {
		return errMsg{err}
	}
	return nil
}

func (m *Model) handleApplyStash() tea.Msg {
	if err := git.StashApply(m.dir, m.stashes[m.panels[focusStash].idx].Ref); err != nil {
		return errMsg{err}
	}
	files, err := git.Status(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return filesMsg(files)
}

func (m *Model) handleDropStash() tea.Msg {
	if err := git.StashDrop(m.dir, m.stashes[m.panels[focusStash].idx].Ref); err != nil {
		return errMsg{err}
	}
	stashes, err := git.Stashes(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return stashesMsg(stashes)
}

func (m *Model) handlePopStash() tea.Msg {
	if err := git.StashPop(m.dir, m.stashes[m.panels[focusStash].idx].Ref); err != nil {
		return errMsg{err}
	}
	stashes, err := git.Stashes(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return stashesMsg(stashes)
}

func (m *Model) handleClearStash() tea.Msg {
	if err := git.StashClear(m.dir); err != nil {
		return errMsg{err}
	}
	stashes, err := git.Stashes(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return stashesMsg(stashes)
}

func (m *Model) handleDropCommit() tea.Msg {
	if err := git.DropLastCommit(m.dir); err != nil {
		return errMsg{err}
	}
	log, err := git.Log(m.dir, "HEAD", 200)
	if err != nil {
		return errMsg{err}
	}
	return logMsg(log)
}

func (m *Model) handlePushStash() tea.Msg {
	if err := git.StashPush(m.dir, m.inputPopup.input.Value()); err != nil {
		return errMsg{err}
	}
	stashes, err := git.Stashes(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return stashesMsg(stashes)
}

func (m *Model) handleStashBranch() tea.Msg {
	branch := m.stashBranchPopup.input.Value()
	if err := git.StashBranch(m.dir, branch, m.stashes[m.panels[focusStash].idx].Ref); err != nil {
		return errMsg{err}
	}
	return nil
}

func (m *Model) handleRenameBranch(oldName string) tea.Msg {
	if err := git.RenameBranch(m.dir, oldName, m.inputPopup.input.Value()); err != nil {
		return errMsg{err}
	}
	branches, err := git.Branches(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return branchesMsg(branches)
}

func (m *Model) handleSetRemote() tea.Msg {
	if err := git.SetRemote(m.dir, m.inputPopup.input.Value()); err != nil {
		return errMsg{err}
	}
	return nil
}

func (m *Model) handleCreateBranch() tea.Msg {
	if err := git.CreateBranch(m.dir, m.inputPopup.input.Value()); err != nil {
		return errMsg{err}
	}
	if err := git.Checkout(m.dir, m.inputPopup.input.Value()); err != nil {
		return errMsg{err}
	}
	branches, err := git.Branches(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return branchesMsg(branches)
}

func (m *Model) handleInitRepo(branch string) tea.Msg {
	if err := git.Init(m.dir, branch); err != nil {
		return errMsg{err}
	}
	return repoReadyMsg{}
}

func (m *Model) handleToggleStage() tea.Msg {
	idx, ok := m.selectedFile()
	if !ok {
		return nil
	}
	if m.files[idx].Staged() {
		if err := git.Reset(m.dir, m.files[idx].Path); err != nil {
			return errMsg{err}
		}
	} else if err := git.Add(m.dir, m.files[idx].Path); err != nil {
		return errMsg{err}
	}
	files, err := git.Status(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return filesMsg(files)
}

func (m *Model) handleToggleStageAll() tea.Msg {
	if git.AllStaged(m.files) {
		if err := git.ResetAll(m.dir); err != nil {
			return errMsg{err}
		}
	} else if err := git.AddAll(m.dir); err != nil {
		return errMsg{err}
	}
	files, err := git.Status(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return filesMsg(files)
}

func (m *Model) handleAmend() tea.Msg {
	if err := git.Ammend(m.dir, m.log[0].Subject); err != nil {
		return errMsg{err}
	}
	files, err := git.Status(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return filesMsg(files)
}

func (m *Model) handleCommit() tea.Msg {
	summary := m.commitPopup.commitSummary.Value()
	message := m.commitPopup.commitMessage.Value()
	// Git splits subject from body on a blank line: %s stops at the first
	// blank line and folds anything before it onto one line, so a single
	// "\n" here would merge the body into the subject instead of keeping
	// them separate.
	full := summary
	if message != "" {
		full = summary + "\n\n" + message
	}

	var err error
	switch {
	case m.commitPopup.reword && len(m.log) > 0 && m.commitPopup.rewordHash == m.log[0].Hash:
		// Rewording HEAD is a plain amend: no rebase, and it doesn't care
		// about the working tree being dirty the way rebase would.
		err = git.Ammend(m.dir, full)
	case m.commitPopup.squash:
		err = git.SquashCommits(m.dir, m.commitPopup.squashOldestHash, m.commitPopup.squashCount, full)
	case m.commitPopup.reword:
		err = git.RewordCommit(m.dir, m.commitPopup.rewordHash, full)
	default:
		err = git.Commit(m.dir, full)
	}
	if err != nil {
		return errMsg{err}
	}
	return nil
}

func (m *Model) handleTag() tea.Msg {
	name := m.tagPopup.tagName.Value()
	message := m.tagPopup.tagMessage.Value()
	hash := m.tagPopup.hash
	err := git.CreateTag(m.dir, name, message, hash)
	if err != nil {
		return errMsg{err}
	}
	return nil
}

func (m *Model) handleMerge(branch string, mode git.MergeMode) tea.Msg {
	err := git.Merge(m.dir, branch, mode)
	if err != nil {
		return errMsg{err}
	}
	branches, err := git.Branches(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return branchesMsg(branches)
}

func (m *Model) handlePush(branch string) tea.Msg {
	err := git.Push(m.dir, branch)
	if err != nil {
		return errMsg{err}
	}
	branches, err := git.Branches(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return branchesMsg(branches)
}

func (m *Model) handleRewordCommit() tea.Cmd {
	entry := m.log[m.panels[focusLog].idx]
	m.commitPopup.reword = true
	m.commitPopup.rewordHash = entry.Hash
	m.commitPopup.squash = false
	m.commitPopup.squashOldestHash = ""
	m.commitPopup.squashCount = 0
	m.commitPopup.focus = commitFocusSummary
	m.commitPopup.commitSummary.SetWidth(m.width / 3)
	m.commitPopup.commitMessage.SetWidth(m.width / 3)
	m.commitPopup.commitSummary.SetValue(entry.Subject)
	m.commitPopup.commitMessage.SetValue(entry.Body)
	m.commitPopup.commitSummary.Placeholder = "Commit summary"
	m.commitPopup.commitMessage.Placeholder = "Commit message"
	m.commitPopup.commitSummary.Focus()
	m.commitPopup.commitMessage.Blur()
	m.commitPopup.active = true
	return textinput.Blink
}

// startSquash opens commitPopup to confirm squashing the marked log range
// [lo, hi] (indices into m.log, which is newest-first) into one commit,
// pre-filled with the newest subject as the summary and every subject in
// the range, oldest first, as a bulleted message body - mirroring
// handleRewordCommit's single-entry prefill.
func (m *Model) startSquash(lo, hi int) tea.Cmd {
	m.commitPopup.reword = false
	m.commitPopup.rewordHash = ""
	m.commitPopup.squash = true
	m.commitPopup.squashOldestHash = m.log[hi].Hash
	m.commitPopup.squashCount = hi - lo + 1

	subjects := make([]string, 0, hi-lo+1)
	for i := hi; i >= lo; i-- {
		subjects = append(subjects, "- "+m.log[i].Subject)
	}

	m.commitPopup.focus = commitFocusSummary
	m.commitPopup.commitSummary.SetWidth(m.width / 3)
	m.commitPopup.commitMessage.SetWidth(m.width / 3)
	m.commitPopup.commitSummary.SetValue(m.log[lo].Subject)
	m.commitPopup.commitMessage.SetValue(strings.Join(subjects, "\n"))
	m.commitPopup.commitSummary.Placeholder = "Commit summary"
	m.commitPopup.commitMessage.Placeholder = "Commit message"
	m.commitPopup.commitSummary.Focus()
	m.commitPopup.commitMessage.Blur()
	m.commitPopup.active = true
	return textinput.Blink
}

// mouseClick focuses whichever panel a click landed in (via panelAt) and,
// for the four list panels, selects the row under the cursor - the mouse
// equivalent of pressing a focus-switch key followed by moving the cursor
// to that row.
func (m *Model) mouseClick(ms tea.Mouse) {
	target, row, ok := m.panelAt(ms.X, ms.Y)
	if !ok {
		return
	}
	m.focus = target
	if row < 0 || target == focusDiff {
		return
	}
	m.panels[target].idx = row
}

// mouseWheel scrolls the diff viewport when the wheel is over the diff
// panel, or moves the selection cursor by one (like k/j) when it's over one
// of the four list panels - and focuses whichever panel that is, same as
// mouseClick, so the diff pane updates to match what the wheel just moved.
func (m *Model) mouseWheel(ms tea.Mouse) tea.Cmd {
	var delta int
	switch ms.Button {
	case tea.MouseWheelUp:
		delta = -1
	case tea.MouseWheelDown:
		delta = 1
	default:
		return nil
	}

	target, _, ok := m.panelAt(ms.X, ms.Y)
	if !ok {
		return nil
	}
	m.focus = target
	if target == focusDiff {
		if delta < 0 {
			m.diff.ScrollUp(3)
		} else {
			m.diff.ScrollDown(3)
		}
		return nil
	}
	m.move(target, delta)
	return m.showDiff()
}
