package ui

import (
	"bubblegit/internal/git"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case filesMsg:
		m.files = []git.FileStatus(msg)
		return m, nil

	case branchesMsg:
		m.branches = []git.BranchInfo(msg)
		return m, nil

	case tea.WindowSizeMsg:
		m.diff.SetHeight(msg.Height)
		m.diff.SetWidth(msg.Width / 2)
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		m.err = nil
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			m.focus = (m.focus + 1) % focusCount

		case "up":
			switch m.focus {
			case focusBranch:
				m.moveBranch(-1)
			case focusStag:
				m.moveFile(-1)
			}

		case "down":
			switch m.focus {
			case focusBranch:
				m.moveBranch(1)
			case focusStag:
				m.moveFile(1)
			}

		case "d":
			if m.focus == focusBranch && len(m.branches) > 0 {
				branch := m.branches[m.branchFocus].Name
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
				old := m.branches[m.branchFocus].Name
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
					err := git.Checkout(m.dir, m.branches[m.branchFocus].Name)
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
	}

	return m, nil
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
	m.branchFocus += delta

	if m.branchFocus < 0 {
		m.branchFocus = 0
	}

	if m.branchFocus >= len(m.branches) {
		m.branchFocus = len(m.branches) - 1
	}
}
