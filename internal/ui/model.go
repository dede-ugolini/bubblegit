// Package ui provides the ui for application
package ui

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	dir string

	branch string

	diff viewport.Model

	ready    bool
	quitting bool
}

func NewModel(dir string) Model {
	return Model{
		dir: dir,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
