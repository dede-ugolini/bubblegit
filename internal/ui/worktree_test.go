package ui

import (
	"strings"
	"testing"

	"bubblegit/internal/git"
)

func TestRenderWorktreePopup(t *testing.T) {
	m := NewModel(".")
	m.width, m.height = 100, 40
	m.worktreePopup.active = true
	m.worktreePopup.path.Prompt = "path> "
	m.worktreePopup.branch.Prompt = "branch> "
	m.worktreePopup.path.Focus()

	out := m.renderWorktreePopup()
	for _, want := range []string{"New worktree", "path> ", "branch> ", "ctrl+s add", "cancel"} {
		if !strings.Contains(out, want) {
			t.Errorf("renderWorktreePopup() missing %q in:\n%s", want, out)
		}
	}
}

func TestRenderWorktrees(t *testing.T) {
	m := NewModel(".")
	m.width, m.height = 100, 40
	m.focus = focusStag
	m.panels[focusStag].view = viewWorktrees
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
	m.panels[focusStag].view = viewWorktrees
	m.setLayout(layoutNormal)

	// An empty list still has to draw its box (same placeholder as stash),
	// or the slot collapses and shifts the panels below it.
	if out := m.renderWorktrees(); !strings.Contains(out, "Worktrees") {
		t.Errorf("renderWorktrees() = %q, want a titled box even when empty", out)
	}
}

// TestCycleView covers the view toggle: switching to the worktrees subpanel
// and back remembers each view's cursor independently.
func TestCycleView(t *testing.T) {
	m := NewModel(".")
	m.focus = focusStag
	m.treeRows = make([]fileRow, 5)
	m.worktrees = make([]git.WorktreeEntry, 3)
	m.panels[focusStag].idx = 3

	m.cycleView()
	if v := m.panels[focusStag].view; v != viewWorktrees {
		t.Fatalf("view = %v, want worktrees", v)
	}
	m.move(focusStag, 1)
	m.move(focusStag, 1)
	if got := m.panels[focusStag].idx; got != 2 {
		t.Fatalf("worktrees cursor = %d, want 2", got)
	}

	m.cycleView()
	if v := m.panels[focusStag].view; v != viewMain {
		t.Fatalf("view = %v, want main", v)
	}
	if got := m.panels[focusStag].idx; got != 3 {
		t.Errorf("files cursor = %d, want 3 after toggling back", got)
	}

	m.cycleView()
	if got := m.panels[focusStag].idx; got != 2 {
		t.Errorf("worktrees cursor = %d, want 2 after toggling back", got)
	}
}

// TestWorktreesGeometry pins down the layout wiring: the worktrees view
// shares the files panel's box while that panel is shown, and collapses
// whenever the files panel does.
func TestWorktreesGeometry(t *testing.T) {
	m := NewModel(".")
	m.width, m.height = 100, 40

	m.focus = focusStag
	m.setLayout(layoutNormal)
	files := m.panels[focusStag]
	if files.height <= 0 || files.width <= 0 {
		t.Fatalf("files geometry = %dx%d, want non-zero under split layout", files.height, files.width)
	}

	// The worktrees view draws into the files panel's own box, so switching
	// to it leaves the geometry untouched.
	m.panels[focusStag].view = viewWorktrees
	if wt := m.panels[focusStag]; wt.height != files.height || wt.width != files.width {
		t.Errorf("worktrees %dx%d, want files %dx%d", wt.height, wt.width, files.height, files.width)
	}

	// Fullscreen on another panel collapses the files panel, and the
	// worktrees view with it.
	m.focus = focusLog
	m.setLayout(layoutFull)
	if wt := m.panels[focusStag]; wt.height != 0 || wt.width != 0 {
		t.Errorf("worktrees geometry = %dx%d, want 0x0 when files is collapsed", wt.height, wt.width)
	}
}
