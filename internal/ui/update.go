package ui

import (
	"bubblegit/internal/git"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.commitPopup.active {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "ctrl+s":
				summary := m.commitPopup.commitSummary.Value()
				if summary == "" {
					return m, nil
				}
				message := m.commitPopup.commitMessage.Value()
				m.commitPopup.commitSummary.Blur()
				m.commitPopup.commitMessage.Blur()
				m.commitPopup.active = false
				return m, tea.Batch(
					func() tea.Msg {
						if err := git.Commit(m.dir, summary+"\n"+message); err != nil {
							return errMsg{err}
						}
						return nil
					},
					m.Refresh(),
				)
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

	if m.focusInput {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter":
				name := m.input.Value()
				old := m.renameBranch
				m.focusInput = false
				m.renameBranch = ""
				if name == "" {
					return m, nil
				}
				return m, func() tea.Msg {
					if old != "" {
						if err := git.RenameBranch(m.dir, old, name); err != nil {
							return errMsg{err}
						}
					} else {
						if err := git.CreateBranch(m.dir, name); err != nil {
							return errMsg{err}
						}
						if err := git.Checkout(m.dir, name); err != nil {
							return errMsg{err}
						}
					}
					branches, err := git.Branches(m.dir)
					if err != nil {
						return errMsg{err}
					}
					return branchesMsg(branches)
				}
			case "esc":
				m.focusInput = false
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
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
		m.filesHeight = msg.Height * 30 / 100
		m.filesWidth = msg.Width * 45 / 100
		m.branchHeight = msg.Height * 15 / 100
		m.branchWidth = msg.Width * 45 / 100
		m.logHeight = msg.Height * 30 / 100
		m.logWidth = msg.Width * 45 / 100
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

		case "0":
			m.focus = focusDiff

		case "1":
			m.focus = focusStag

		case "2":
			m.focus = focusBranch

		case "3":
			m.focus = focusLog

		case "up", "k":
			switch m.focus {
			case focusStag:
				m.moveFile(-1)
			case focusBranch:
				m.moveBranch(-1)
			case focusLog:
				m.moveLog(-1)
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
			case focusDiff:
				var cmd tea.Cmd
				m.diff, cmd = m.diff.Update(msg)
				return m, cmd

			}
			return m, m.showDiff()

		case "d":
			if m.focus == focusBranch && len(m.branches) > 0 {
				branch := m.branches[m.idxBranch].Name
				return m, func() tea.Msg {
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
			}
			if m.focus == focusStag && len(m.branches) > 0 {
				return m, func() tea.Msg {
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
			}

		case "n":
			if m.focus == focusBranch {
				m.input.SetWidth(20)
				m.input.CharLimit = 20
				m.input.SetValue("")
				m.input.Prompt = "branch> "
				m.input.Placeholder = "new branch name"
				m.input.Focus()
				m.focusInput = true
				return m, textinput.Blink
			}

		case "r":
			if m.focus == focusBranch && len(m.branches) > 0 {
				old := m.branches[m.idxBranch].Name
				m.renameBranch = old
				m.input.SetWidth(20)
				m.input.CharLimit = 20
				m.input.SetValue(old)
				m.input.Prompt = "rename> "
				m.input.Placeholder = "new name"
				m.input.Focus()
				m.focusInput = true
				return m, textinput.Blink
			}

		case "enter":
			if m.focus == focusBranch && len(m.branches) > 0 {
				return m, func() tea.Msg {
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
			}
		case "space":
			if m.focus == focusStag && len(m.files) > 0 {
				return m, func() tea.Msg {
					if m.files[m.idxFiles].Staged() {
						err := git.Reset(m.dir, m.files[m.idxFiles].Path)
						if err != nil {
							return errMsg{err: err}
						}
						files, err := git.Status(m.dir)
						if err != nil {
							return errMsg{err: err}
						}
						return filesMsg(files)
					}
					err := git.Add(m.dir, m.files[m.idxFiles].Path)
					if err != nil {
						return errMsg{err: err}
					}
					files, err := git.Status(m.dir)
					if err != nil {
						return errMsg{err: err}
					}
					return filesMsg(files)
				}
			}
		case "a":
			if m.focus == focusStag && len(m.files) > 0 {
				return m, func() tea.Msg {
					if git.AllStaged(m.files) {
						err := git.ResetAll(m.dir)
						if err != nil {
							return errMsg{err}
						}
						files, err := git.Status(m.dir)
						if err != nil {
							return errMsg{err: err}
						}
						return filesMsg(files)

					} else {
						err := git.AddAll(m.dir)
						if err != nil {
							return errMsg{err}
						}
						files, err := git.Status(m.dir)
						if err != nil {
							return errMsg{err: err}
						}
						return filesMsg(files)
					}
				}
			}
		case "c":
			if m.focus == focusStag && len(m.files) > 0 && git.HasOneStaged(m.files) {
				m.commitPopup.focus = commitFocusSummary
				m.commitPopup.commitSummary.SetWidth(m.width / 3)
				m.commitPopup.commitMessage.SetWidth(m.width / 3)
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

					m.diff.SetHeight(0)
					m.diff.SetWidth(0)
				case focusBranch:
					m.branchHeight = m.height
					m.branchWidth = m.width

					m.filesHeight = 0
					m.filesWidth = 0

					m.logHeight = 0
					m.logWidth = 0

					m.diff.SetHeight(0)
					m.diff.SetWidth(0)
				case focusLog:
					m.logHeight = m.height
					m.logWidth = m.width

					m.filesHeight = 0
					m.filesWidth = 0

					m.branchHeight = 0
					m.branchWidth = 0

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
				}
				return m, m.showDiff()
			}

		case "-":
			if m.panelFullScreen {
				m.panelFullScreen = false
				m.filesHeight = m.height * 30 / 100
				m.filesWidth = m.width * 45 / 100

				m.branchHeight = m.height * 15 / 100
				m.branchWidth = m.width * 45 / 100

				m.logHeight = m.height * 30 / 100
				m.logWidth = m.width * 45 / 100

				m.diff.SetHeight(m.height * 90 / 100)
				m.diff.SetWidth(m.width * 55 / 100)
				return m, m.showDiff()
			}

		case "pgup":
			m.diff.ScrollUp(8)
		case "pgdown":
			m.diff.ScrollDown(8)
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
			diff, err := git.DiffDelta(m.dir, m.files[m.idxFiles].Path, m.panelFullScreen, m.diff.Width())
			if err != nil {
				return errMsg{err}
			}
			return diffMsg{diff}
		case focusBranch:
			if len(m.branches) <= 0 {
				return diffMsg{}
			}
			diff, err := git.DiffBranchDelta(m.dir, m.panelFullScreen, m.diff.Width())
			if err != nil {
				return errMsg{err}
			}
			return diffMsg{diff}
		case focusLog:
			if len(m.log) == 0 {
				return diffMsg{}
			}
			diff, err := git.ShowDelta(m.dir, m.log[m.idxLog].Hash, m.panelFullScreen, m.diff.Width())
			if err != nil {
				return errMsg{err}
			}
			return diffMsg{diff}
		}
		return nil
	}
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
