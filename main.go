package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

type model struct {
	dir string

	branch string

	diff viewport.Model

	ready    bool
	quitting bool
}

func initialModel(dir string) model {
	return model{
		dir:  dir,
		diff: viewport.New(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.diff.SetHeight(msg.Height)
		m.diff.SetWidth(msg.Width / 2)
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	if m.quitting {
		tea.NewView("")
	}

	// TODO: add spinner
	if !m.ready {
		tea.NewView("Loading...")
	}

	branch, _ := Branch()

	return tea.NewView(branch)
	/*
		return tea.NewView(fmt.Sprintf(
			"Olá, Bubble Tea!\n\n" +
				"Pressione q para sair.\n",
		))
	*/
}

func Branch() (string, error) {
	out, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "bubblegit:", err)
		os.Exit(1)
	}
	if len(os.Args) < 1 {
		dir = os.Args[0]
	}
	p := tea.NewProgram(initialModel(dir))

	if _, err := p.Run(); err != nil {
		fmt.Println("Erro:", err)
		os.Exit(1)
	}
}
