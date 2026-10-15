package ui

import (
	"strings"
	"testing"

	"bubblegit/internal/git"
)

func TestRenderTags(t *testing.T) {
	m := NewModel(".")
	m.width, m.height = 100, 40
	m.focus = focusBranch
	m.panels[focusBranch].view = viewTags
	m.setLayout(layoutNormal)

	m.tagList = []git.TagEntry{
		{Name: "v1.0", Hash: "abcdef1234567890", Message: "release 1.0"},
		{Name: "light", Hash: "0987654321", Message: "second commit"},
	}

	out := m.renderTags()
	if out == "" {
		t.Fatal("renderTags() = empty, want a rendered panel")
	}
	for _, want := range []string{"Tags", "v1.0", "release 1.0", "light", "second commit"} {
		if !strings.Contains(out, want) {
			t.Errorf("renderTags() missing %q in:\n%s", want, out)
		}
	}
}

func TestRenderTagsEmpty(t *testing.T) {
	m := NewModel(".")
	m.width, m.height = 100, 40
	m.focus = focusBranch
	m.panels[focusBranch].view = viewTags
	m.setLayout(layoutNormal)

	// An empty list still has to draw its box (same placeholder as stash),
	// or the slot collapses and shifts the panels below it.
	if out := m.renderTags(); !strings.Contains(out, "Tags") {
		t.Errorf("renderTags() = %q, want a titled box even when empty", out)
	}
}

// TestCycleViewBranches covers the branch panel's view toggle: switching to
// the tags subpanel and back remembers each view's cursor independently.
func TestCycleViewBranches(t *testing.T) {
	m := NewModel(".")
	m.focus = focusBranch
	m.branches = make([]git.BranchInfo, 5)
	m.tagList = make([]git.TagEntry, 3)
	m.panels[focusBranch].idx = 3

	m.cycleView()
	if v := m.panels[focusBranch].view; v != viewTags {
		t.Fatalf("view = %v, want tags", v)
	}
	m.move(focusBranch, 1)
	m.move(focusBranch, 1)
	if got := m.panels[focusBranch].idx; got != 2 {
		t.Fatalf("tags cursor = %d, want 2", got)
	}

	m.cycleView()
	if v := m.panels[focusBranch].view; v != viewMain {
		t.Fatalf("view = %v, want main", v)
	}
	if got := m.panels[focusBranch].idx; got != 3 {
		t.Errorf("branches cursor = %d, want 3 after toggling back", got)
	}

	m.cycleView()
	if got := m.panels[focusBranch].idx; got != 2 {
		t.Errorf("tags cursor = %d, want 2 after toggling back", got)
	}
}

// TestTagsGeometry pins down the layout wiring: the tags view shares the
// branch panel's box while that panel is shown, and collapses whenever the
// branch panel does.
func TestTagsGeometry(t *testing.T) {
	m := NewModel(".")
	m.width, m.height = 100, 40

	m.focus = focusBranch
	m.setLayout(layoutNormal)
	branches := m.panels[focusBranch]
	if branches.height <= 0 || branches.width <= 0 {
		t.Fatalf("branch geometry = %dx%d, want non-zero under split layout", branches.height, branches.width)
	}

	// The tags view draws into the branch panel's own box, so switching to
	// it leaves the geometry untouched.
	m.panels[focusBranch].view = viewTags
	if tags := m.panels[focusBranch]; tags.height != branches.height || tags.width != branches.width {
		t.Errorf("tags %dx%d, want branch %dx%d", tags.height, tags.width, branches.height, branches.width)
	}

	// Fullscreen on another panel collapses the branch panel, and the tags
	// view with it.
	m.focus = focusLog
	m.setLayout(layoutFull)
	if tags := m.panels[focusBranch]; tags.height != 0 || tags.width != 0 {
		t.Errorf("tags geometry = %dx%d, want 0x0 when branch is collapsed", tags.height, tags.width)
	}
}
