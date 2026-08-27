// Package git provides git operations
package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// FileStatus represents one line of `git status --porcelain` output.
type FileStatus struct {
	Index    byte
	Worktree byte
	Path     string

	OrigPath string
}

// LogEntry is one commit as reported by `git log`.
type LogEntry struct {
	Hash      string
	ShortHash string
	Author    string
	Date      string
	Subject   string
}

// field/record separators unlikely to appear in commit metadata.
const (
	logFieldSep  = "\x1f"
	logRecordSep = "\x1e"
)

// TODO: Refactor to avoid always using string(out)

// Status returns the working tree status
func Status(dir string) ([]FileStatus, error) {
	cmd := exec.Command(
		"git",
		"status",
		"--porcelain",
		"-z",
		"--untracked-files=all",
	)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	if string(out) == "" {
		return nil, nil
	}

	return parseStatus(string(out))
}

func parseStatus(out string) ([]FileStatus, error) {
	var files []FileStatus
	parts := strings.Split(strings.TrimRight(out, "\x00"), "\x00")

	for i := 0; i < len(parts); i++ {
		entry := parts[i]
		if len(entry) < 3 {
			continue
		}
		fs := FileStatus{
			Index:    entry[0],
			Worktree: entry[1],
			Path:     entry[3:],
		}
		// Renames/copies have an extra NUL-separated original path.
		if fs.Index == 'R' || fs.Index == 'C' {
			i++
			if i < len(parts) {
				fs.OrigPath = parts[i]
			}
		}
		files = append(files, fs)
	}
	return files, nil
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

func Restore(dir, path string) error {
	cmd := exec.Command("git", "restore", "--", path)
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

type BranchInfo struct {
	Name    string
	Current bool
}

// Branches lists local branches, current branch first.
func Branches(dir string) ([]BranchInfo, error) {
	cmd := exec.Command("git", "branch", "--format=%(HEAD)"+logFieldSep+"%(refname)"+logFieldSep+"%(refname:short)")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var branches []BranchInfo
	for line := range strings.SplitSeq(string(out), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		f := strings.SplitN(line, logFieldSep, 3)
		if len(f) != 3 {
			continue
		}
		// In detached-HEAD state, git lists a synthetic
		// "(HEAD detached at ...)" row whose refname isn't a real ref
		// under refs/heads/ — skip it, it's not a checkout/delete target.
		if !strings.HasPrefix(f[1], "refs/heads/") {
			continue
		}
		branches = append(branches, BranchInfo{Name: f[2], Current: f[0] == "*"})
	}
	sort.SliceStable(branches, func(i, j int) bool {
		return branches[i].Current && !branches[j].Current
	})
	return branches, nil
}

// Checkout switches the working tree to the given branch.
func Checkout(dir, branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// CreateBranch create and checks out a new branch off the current HEAD.
func CreateBranch(dir, branch string) error {
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// DeleteBranch removes a local branch. It refuses (like plain `git branch
// -d`) if the branch has commits not merged elsewhere.
func DeleteBranch(dir, branch string) error {
	cmd := exec.Command("git", "branch", "-d", branch)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// RenameBranch rename a branch
func RenameBranch(dir, oldName, newName string) error {
	cmd := exec.Command("git", "branch", "-m", oldName, newName)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func Log(dir, rev string, limit int) ([]LogEntry, error) {
	format := strings.Join(
		[]string{"%H", "%h", "%an", "%ad", "%s"}, logFieldSep,
	) + logRecordSep
	cmd := exec.Command(
		"git",
		"log",
		rev,
		"--date=short",
		"--pretty=format:"+format,
		fmt.Sprintf("-n%d", limit),
	)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(err.Error(), "does not have any commits yet") ||
			strings.Contains(err.Error(), "unknown revision") {
			return nil, nil
		}
		return nil, err
	}
	if string(out) == "" {
		return nil, nil
	}
	var entries []LogEntry
	for rec := range strings.SplitSeq(string(out), logRecordSep) {
		rec = strings.TrimPrefix(rec, "\n")
		if rec == "" {
			continue
		}
		f := strings.Split(rec, logFieldSep)
		if len(f) < 5 {
			continue
		}
		entries = append(entries, LogEntry{
			Hash:      f[0],
			ShortHash: f[1],
			Author:    f[2],
			Date:      f[3],
			Subject:   f[4],
		})
	}
	return entries, nil
}

func Diff(dir, path string) (string, error) {
	cmd := exec.Command("git", "diff", "--color=always", "--", path)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(""), err
	}
	return string(out), nil
}

func DiffDelta(dir, path string, sideBySide bool, width int) (string, error) {
	git := exec.Command("git", "diff", "--no-color", "--", path)
	git.Dir = dir
	diff, err := git.Output()
	if err != nil {
		return "", err
	}
	delta := exec.Command(
		"delta",
		"--no-gitconfig",
		"--paging=never",
		"--line-numbers",
	)
	if sideBySide {
		delta.Args = append(delta.Args, "--side-by-side", fmt.Sprintf("--width=%d", width))
	}
	delta.Stdin = bytes.NewReader(diff)

	var output bytes.Buffer
	delta.Stdout = &output

	if err := delta.Run(); err != nil {
		return "", err
	}
	return output.String(), nil
}

func DiffBranchDelta(dir string, sideBySide bool, width int) (string, error) {
	git := exec.Command("git", "diff", "--color=always", "--stat", "--patch")
	git.Dir = dir
	diff, err := git.Output()
	if err != nil {
		return "", err
	}
	delta := exec.Command(
		"delta",
		"--no-gitconfig",
		"--paging=never",
		"--line-numbers",
	)
	if sideBySide {
		delta.Args = append(delta.Args, "--side-by-side", fmt.Sprintf("--width=%d", width))
	}
	delta.Stdin = bytes.NewReader(diff)

	var output bytes.Buffer
	delta.Stdout = &output

	if err := delta.Run(); err != nil {
		return "", err
	}
	return output.String(), nil
}

func Show(dir, hash string) (string, error) {
	cmd := exec.Command("git", "show", "--color=always", "--stat", "--patch", hash)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(""), err
	}
	return string(out), nil
}

func ShowDelta(dir, path string, sideBySide bool, width int) (string, error) {
	git := exec.Command("git", "show", "--color=always", "--stat", "--patch", path)
	git.Dir = dir
	diff, err := git.Output()
	if err != nil {
		return "", err
	}
	delta := exec.Command(
		"delta",
		"--no-gitconfig",
		"--paging=never",
		"--line-numbers",
	)
	if sideBySide {
		delta.Args = append(delta.Args, "--side-by-side", fmt.Sprintf("--width=%d", width))
	}
	delta.Stdin = bytes.NewReader(diff)

	var output bytes.Buffer
	delta.Stdout = &output

	if err := delta.Run(); err != nil {
		return "", err
	}
	return output.String(), nil
}

func Commit(dir, message string) {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	cmd.Run()
}
