package ui

import (
	"bubblegit/internal/git"

	tea "charm.land/bubbletea/v2"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case errMsg:
		m.err = msg.err
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
			}

		case "down":
			switch m.focus {
			case focusBranch:
				m.moveBranch(1)
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
		}
	}

	return m, nil
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
