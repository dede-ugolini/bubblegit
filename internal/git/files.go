package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// FileStatus represents one line of `git status --porcelain` output.
type FileStatus struct {
	Index    byte
	Worktree byte
	Path     string

	OrigPath string
}

// Add stages the given path
func Add(dir, path string) error {
	cmd := exec.Command("git", "add", "--", path)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// AddAll stages all path
func AddAll(dir string) error {
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// Reset unstages the given path, leaving working tree changes intact.
func Reset(dir, path string) error {
	cmd := exec.Command("git", "reset", "--", path)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func ResetAll(dir string) error {
	cmd := exec.Command("git", "reset")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// Restore reverts the file to its HEAD state, discarding both staged and
// unstaged changes.
func Restore(dir, path string) error {
	cmd := exec.Command("git", "restore", "--staged", "--worktree", "--", path)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func RestoreUntracked(dir, path string) error {
	cmd := exec.Command("git", "clean", "-fd", path)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// Untracked reports whether this file is untracked by git.
func (f FileStatus) Untracked() bool {
	return f.Index == '?' && f.Worktree == '?'
}

// Staged reports whether this entry has staged changes to commit.
func (f FileStatus) Staged() bool {
	return f.Index != ' ' && f.Index != '?'
}

// Unstaged reports whether this entry has changes not yet staged
// (including being an untracked file).
func (f FileStatus) Unstaged() bool {
	return f.Worktree != ' ' || (f.Index == '?' && f.Worktree == '?')
}

// AllStaged reports wheter all entrys
func AllStaged(files []FileStatus) bool {
	for _, f := range files {
		if !f.Staged() {
			return false
		}
	}
	return true
}

func HasOneStaged(files []FileStatus) bool {
	for _, f := range files {
		if f.Staged() {
			return true
		}
	}
	return false
}
