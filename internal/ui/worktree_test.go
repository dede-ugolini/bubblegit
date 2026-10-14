package ui

import (
	"strings"
	"testing"

	"bubblegit/internal/git"
)

func TestRenderWorktrees(t *testing.T) {
	m := NewModel(".")
	m.width, m.height = 100, 40
	m.focus = focusStag
	m.subPanelFocus = subPanelFocusWorktrees
	m.setLayout(layoutNormal)

	m.worktrees = []git.WorktreeEntry{
		{Path: "/repo", Hash: "abcdef1234567890", ShortHash: "abcdef1", Branch: "main"},
		{Path: "/repo-linked", Hash: "0987654321", ShortHash: "0987654", Branch: "(detached)"},
	}

	out := m.renderWorktrees()
	if out == "" {
		t.Fatal("renderWorktrees() = empty, want a rendered panel")
	}
	for _, want := range []string{"Worktrees", "abcdef1", "main", "0987654", "(detached)", "/repo-linked"} {
		if !strings.Contains(out, want) {
			t.Errorf("renderWorktrees() missing %q in:\n%s", want, out)
		}
	}
}

func TestRenderWorktreesEmpty(t *testing.T) {
	m := NewModel(".")
	m.width, m.height = 100, 40
	m.focus = focusStag
	m.subPanelFocus = subPanelFocusWorktrees
	m.setLayout(layoutNormal)

	// An empty list still has to draw its box (same placeholder as stash),
	// or the slot collapses and shifts the panels below it.
	if out := m.renderWorktrees(); !strings.Contains(out, "Worktrees") {
		t.Errorf("renderWorktrees() = %q, want a titled box even when empty", out)
	}
}

// TestWorktreesGeometry pins down the layout wiring: the worktrees subpanel
// shares the files panel's box while that panel is shown, and collapses
// whenever the files panel does.
func TestWorktreesGeometry(t *testing.T) {
	m := NewModel(".")
	m.width, m.height = 100, 40

	m.focus = focusStag
	m.setLayout(layoutNormal)
	wt := m.subPanels[subPanelFocusWorktrees]
	files := m.panels[focusStag]
	if wt.height != files.height || wt.width != files.width {
		t.Errorf("worktrees %dx%d, want files %dx%d", wt.height, wt.width, files.height, files.width)
	}
	if wt.height <= 0 || wt.width <= 0 {
		t.Fatalf("worktrees geometry = %dx%d, want non-zero under split layout", wt.height, wt.width)
	}

	// Fullscreen on another panel collapses the files panel, and the
	// worktrees subpanel with it.
	m.focus = focusLog
	m.setLayout(layoutFull)
	if wt := m.subPanels[subPanelFocusWorktrees]; wt.height != 0 || wt.width != 0 {
		t.Errorf("worktrees geometry = %dx%d, want 0x0 when files is collapsed", wt.height, wt.width)
	}
}
