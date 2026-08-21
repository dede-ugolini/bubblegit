package ui

import (
	"bubblegit/internal/git"

	tea "charm.land/bubbletea/v2"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case branchesMsg:
		m.branches = []git.BranchInfo(msg)
		return m, nil

	case tea.WindowSizeMsg:
		m.diff.SetHeight(msg.Height)
		m.diff.SetWidth(msg.Width / 2)
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			m.focus = (m.focus + 1) % focusCount

		case "up":
			if m.branchFocus > 0 {
				m.branchFocus--
			}
		case "down":
			if m.branchFocus < len(m.branches)-1 {
				m.branchFocus++
			}
		}
	}

	return m, nil
}
