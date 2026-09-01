package ui

import (
	"fmt"
	"strings"

	"bubblegit/internal/git"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.squashMarking {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "up", "k":
				m.moveLog(-1)
				return m, m.showDiff()
			case "down", "j":
				m.moveLog(1)
				return m, m.showDiff()
			case "S":
				lo, hi := m.squashAnchor, m.idxLog
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

	if m.stashClearConfirm {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "y", "enter":
				m.stashClearConfirm = false
				return m, m.handleClearStash
			case "n", "esc":
				m.stashClearConfirm = false
				return m, nil
			}
		}
		return m, nil
	}

	if m.dropConfirm {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "y", "enter":
				m.dropConfirm = false
				return m, m.handleDropCommit
			case "n", "esc":
				m.dropConfirm = false
				return m, nil
			}
		}
		return m, nil
	}

	if m.mergePopup.active {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "up", "k":
				m.mergePopup.idx = (m.mergePopup.idx - 1 + len(mergeModes)) % len(mergeModes)
				return m, nil
			case "down", "j":
				m.mergePopup.idx = (m.mergePopup.idx + 1) % len(mergeModes)
				return m, nil
			case "enter":
				mode := mergeModes[m.mergePopup.idx]
				branch := m.mergePopup.branch
				m.mergePopup.active = false
				return m, tea.Batch(func() tea.Msg { return m.handleMerge(branch, mode) }, m.Refresh())
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
				return m, tea.Batch(m.handleStashBranch, m.Refresh())
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
				return m, tea.Batch(m.handleCommit, m.Refresh())
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
				return m, tea.Batch(m.handleTag, m.Refresh())
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
				}
				return m, nil
			case "esc":
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

	case tickMsg:
		return m, tea.Batch(m.Refresh(), tickCmd())

	case filesMsg:
		m.files = []git.FileStatus(msg)
		return m, nil

	case branchesMsg:
		m.branches = []git.BranchInfo(msg)
		return m, nil

	case logMsg:
		m.log = []git.LogEntry(msg)
		return m, nil

	case stashesMsg:
		m.stashes = []git.StashEntry(msg)
		if m.idxStash >= len(m.stashes) {
			m.idxStash = len(m.stashes) - 1
		}
		if m.idxStash < 0 {
			m.idxStash = 0
		}
		return m, nil

	case diffMsg:
		m.diff.SetContent(msg.diff)
		return m, nil

	case tea.MouseWheelMsg:
		var cmd tea.Cmd
		m.diff, cmd = m.diff.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.filesHeight = msg.Height * 25 / 100
		m.filesWidth = msg.Width * 45 / 100
		m.branchHeight = msg.Height * 15 / 100
		m.branchWidth = msg.Width * 45 / 100
		m.logHeight = msg.Height * 20 / 100
		m.logWidth = msg.Width * 45 / 100
		m.stashHeight = msg.Height * 15 / 100
		m.stashWidth = msg.Width * 45 / 100
		m.diff.SetHeight(msg.Height * 90 / 100)
		m.diff.SetWidth(msg.Width * 55 / 100)
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
			switch m.focus {
			case focusStag:
				m.moveFile(-1)
			case focusBranch:
				m.moveBranch(-1)
			case focusLog:
				m.moveLog(-1)
			case focusStash:
				m.moveStash(-1)
			}
			return m, m.showDiff()

		case "down", "j":
			switch m.focus {
			case focusStag:
				m.moveFile(1)
			case focusBranch:
				m.moveBranch(1)
			case focusLog:
				m.moveLog(1)
			case focusStash:
				m.moveStash(1)
			case focusDiff:
				var cmd tea.Cmd
				m.diff, cmd = m.diff.Update(msg)
				return m, cmd

			}
			return m, m.showDiff()

		case "d":
			// Delete Branch
			if m.focus == focusBranch && len(m.branches) > 0 {
				return m, m.handleDeleteBranch
			}
			// Restore file
			if m.focus == focusStag && len(m.files) > 0 {
				return m, m.handleRestoreFile
			}
			// Drop stash
			if m.focus == focusStash && len(m.stashes) > 0 {
				return m, m.handleDropStash
			}
			// Drop last commit
			if m.focus == focusLog && len(m.log) > 0 {
				m.dropConfirm = true
				m.dropSubject = m.log[0].Subject
			}

		case "D":
			// Clear Stash
			if m.focus == focusStash && len(m.stashes) > 0 {
				m.stashClearConfirm = true
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
				old := m.branches[m.idxBranch].Name
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
			// Checkout Branch
			if m.focus == focusBranch && len(m.branches) > 0 {
				return m, m.handleCheckoutBranch
			}
			// Apply Stash
			if m.focus == focusStash && len(m.stashes) > 0 {
				return m, m.handleApplyStash
			}
		case "space":
			// stage/unstage file
			if m.focus == focusStag && len(m.files) > 0 {
				return m, m.handleToggleStage
			}
		case "a":
			// stage/unstage all
			if m.focus == focusStag && len(m.files) > 0 {
				return m, m.handleToggleStageAll
			}
		case "A":
			// Amend commit
			if m.focus == focusStag && len(m.files) > 0 && git.HasOneStaged(m.files) {
				return m, m.handleAmend
			}
		case "p":
			// Pop stash
			if m.focus == focusStash && len(m.stashes) > 0 {
				return m, m.handlePopStash
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

		case "+":
			if !m.panelFullScreen {
				m.panelFullScreen = true
				switch m.focus {
				case focusStag:
					m.filesHeight = m.height
					m.filesWidth = m.width

					m.branchHeight = 0
					m.branchWidth = 0

					m.logHeight = 0
					m.logWidth = 0

					m.stashHeight = 0
					m.stashWidth = 0

					m.diff.SetHeight(0)
					m.diff.SetWidth(0)
				case focusBranch:
					m.branchHeight = m.height
					m.branchWidth = m.width

					m.filesHeight = 0
					m.filesWidth = 0

					m.logHeight = 0
					m.logWidth = 0

					m.stashHeight = 0
					m.stashWidth = 0

					m.diff.SetHeight(0)
					m.diff.SetWidth(0)
				case focusLog:
					m.logHeight = m.height
					m.logWidth = m.width

					m.filesHeight = 0
					m.filesWidth = 0

					m.branchHeight = 0
					m.branchWidth = 0

					m.stashHeight = 0
					m.stashWidth = 0

					m.diff.SetHeight(0)
					m.diff.SetWidth(0)
				case focusStash:
					m.stashHeight = m.height
					m.stashWidth = m.width

					m.filesHeight = 0
					m.filesWidth = 0

					m.branchHeight = 0
					m.branchWidth = 0

					m.logHeight = 0
					m.logWidth = 0

					m.diff.SetHeight(0)
					m.diff.SetWidth(0)
				case focusDiff:
					m.diff.SetHeight(m.height)
					m.diff.SetWidth(m.width)

					m.filesHeight = 0
					m.filesWidth = 0

					m.branchHeight = 0
					m.branchWidth = 0

					m.logHeight = 0
					m.logWidth = 0

					m.stashHeight = 0
					m.stashWidth = 0
				}
				return m, m.showDiff()
			}

		case "-":
			if m.panelFullScreen {
				m.panelFullScreen = false
				m.filesHeight = m.height * 25 / 100
				m.filesWidth = m.width * 45 / 100

				m.branchHeight = m.height * 15 / 100
				m.branchWidth = m.width * 45 / 100

				m.logHeight = m.height * 20 / 100
				m.logWidth = m.width * 45 / 100

				m.stashHeight = m.height * 15 / 100
				m.stashWidth = m.width * 45 / 100

				m.diff.SetHeight(m.height * 90 / 100)
				m.diff.SetWidth(m.width * 55 / 100)
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
				m.squashAnchor = m.idxLog
			}

		case "M":
			if m.focus == focusBranch && len(m.branches) > 0 {
				m.mergePopup.active = true
				m.mergePopup.idx = 0
				m.mergePopup.branch = m.branches[m.idxBranch].Name
			}

		case "P":
			if m.focus == focusBranch && len(m.branches) > 0 {
				branch := m.branches[m.idxBranch].Name
				return m, tea.Batch(func() tea.Msg { return m.handlePush(branch) }, m.Refresh())
			}

		case "T":
			if m.focus == focusLog && len(m.log) > 0 {
				m.tagPopup.active = true
				m.tagPopup.hash = m.log[m.idxLog].Hash
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

func (m Model) showDiff() tea.Cmd {
	return func() tea.Msg {
		switch m.focus {
		case focusStag:
			if len(m.files) <= 0 {
				return diffMsg{}
			}
			var (
				diff string
				err  error
			)
			file := m.files[m.idxFiles]
			// Untracked() is also true for Unstaged() (an untracked file has
			// no staged changes, so its worktree side is by definition
			// unstaged) and a partially-staged file satisfies both Staged()
			// and Unstaged() - these are deliberately if/else-if, in
			// priority order, so exactly one diff is computed per file.
			if file.Staged() {
				if m.useDelta {
					diff, err = git.DiffDeltaStaged(m.dir, m.files[m.idxFiles].Path, m.panelFullScreen, m.diff.Width())
				} else {
					diff, err = git.DiffStaged(m.dir, m.files[m.idxFiles].Path)
				}
			} else if file.Untracked() {
				if m.useDelta {
					diff, err = git.DiffDeltaUntracked(m.dir, m.files[m.idxFiles].Path, m.panelFullScreen, m.diff.Width())
				} else {
					diff, err = git.DiffUntracked(m.dir, m.files[m.idxFiles].Path)
				}
			} else if file.Unstaged() {
				if m.useDelta {
					diff, err = git.DiffDelta(m.dir, m.files[m.idxFiles].Path, m.panelFullScreen, m.diff.Width())
				} else {
					diff, err = git.Diff(m.dir, m.files[m.idxFiles].Path)
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
				diff, err = git.DiffBranchDelta(m.dir, m.panelFullScreen, m.diff.Width())
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
			if m.useDelta {
				diff, err = git.ShowDelta(m.dir, m.log[m.idxLog].Hash, m.panelFullScreen, m.diff.Width())
			} else {
				diff, err = git.Show(m.dir, m.log[m.idxLog].Hash)
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
			if m.useDelta {
				diff, err = git.StashShowDelta(m.dir, m.stashes[m.idxStash].Ref, m.panelFullScreen, m.diff.Width())
			} else {
				diff, err = git.StashShow(m.dir, m.stashes[m.idxStash].Ref)
			}
			if err != nil {
				return errMsg{err}
			}
			return diffMsg{diff}
		}
		return nil
	}
}

func (m *Model) handleDeleteBranch() tea.Msg {
	branch := m.branches[m.idxBranch].Name
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
	if m.files[m.idxFiles].Untracked() {
		err := git.RestoreUntracked(m.dir, m.files[m.idxFiles].Path)
		if err != nil {
			return errMsg{err}
		}
	} else {
		err := git.Restore(m.dir, m.files[m.idxFiles].Path)
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

func (m *Model) handleCheckoutBranch() tea.Msg {
	err := git.Checkout(m.dir, m.branches[m.idxBranch].Name)
	if err != nil {
		return errMsg{err}
	}
	branches, err := git.Branches(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return branchesMsg(branches)
}

func (m *Model) handleApplyStash() tea.Msg {
	if err := git.StashApply(m.dir, m.stashes[m.idxStash].Ref); err != nil {
		return errMsg{err}
	}
	files, err := git.Status(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return filesMsg(files)
}

func (m *Model) handleDropStash() tea.Msg {
	if err := git.StashDrop(m.dir, m.stashes[m.idxStash].Ref); err != nil {
		return errMsg{err}
	}
	stashes, err := git.Stashes(m.dir)
	if err != nil {
		return errMsg{err}
	}
	return stashesMsg(stashes)
}

func (m *Model) handlePopStash() tea.Msg {
	if err := git.StashPop(m.dir, m.stashes[m.idxStash].Ref); err != nil {
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
	if err := git.StashBranch(m.dir, branch, m.stashes[m.idxStash].Ref); err != nil {
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

func (m *Model) handleToggleStage() tea.Msg {
	if m.files[m.idxFiles].Staged() {
		if err := git.Reset(m.dir, m.files[m.idxFiles].Path); err != nil {
			return errMsg{err}
		}
	} else if err := git.Add(m.dir, m.files[m.idxFiles].Path); err != nil {
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

func (m *Model) handleMerge(branch string, mode mergeMode) tea.Msg {
	err := git.Merge(m.dir, branch, mode.gitArg())
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
	entry := m.log[m.idxLog]
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

func (m *Model) moveFile(delta int) {
	m.idxFiles += delta

	if m.idxFiles < 0 {
		m.idxFiles = 0
	}

	if m.idxFiles >= len(m.files) {
		m.idxFiles = len(m.files) - 1
	}
}

func (m *Model) moveBranch(delta int) {
	m.idxBranch += delta

	if m.idxBranch < 0 {
		m.idxBranch = 0
	}

	if m.idxBranch >= len(m.branches) {
		m.idxBranch = len(m.branches) - 1
	}
}

func (m *Model) moveLog(delta int) {
	m.idxLog += delta
	if m.idxLog < 0 {
		m.idxLog = 0
	}

	if m.idxLog >= len(m.log) {
		m.idxLog = len(m.log) - 1
	}
}

func (m *Model) moveStash(delta int) {
	m.idxStash += delta
	if m.idxStash < 0 {
		m.idxStash = 0
	}

	if m.idxStash >= len(m.stashes) {
		m.idxStash = len(m.stashes) - 1
	}
}
