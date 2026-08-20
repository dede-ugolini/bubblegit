package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"bubblegit/internal/ui"

	tea "charm.land/bubbletea/v2"
)

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
	p := tea.NewProgram(ui.NewModel(dir))

	if _, err := p.Run(); err != nil {
		fmt.Println("Erro:", err)
		os.Exit(1)
	}
}
