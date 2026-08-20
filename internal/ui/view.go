package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func (m Model) View() tea.View {
	if m.quitting {
		tea.NewView("")
	}

	// TODO: add spinner
	if !m.ready {
		tea.NewView("Loading...")
	}

	return tea.NewView(fmt.Sprintf(
		"Olá, Bubble Tea!\n\n" +
			"Pressione q para sair.\n",
	))
}

func renderBranches() {}
