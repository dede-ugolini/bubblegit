// Package git provides git operations
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
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
	Body      string
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
		[]string{"%H", "%h", "%an", "%ad", "%s", "%b"}, logFieldSep,
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
		if len(f) < 6 {
			continue
		}
		entries = append(entries, LogEntry{
			Hash:      f[0],
			ShortHash: f[1],
			Author:    f[2],
			Date:      f[3],
			Subject:   f[4],
			Body:      strings.TrimRight(f[5], "\n"),
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

func DiffBranch(dir string) (string, error) {
	cmd := exec.Command("git", "diff", "--color=always", "--stat", "--patch")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
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

func Commit(dir, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// CreateTag adds an annotated tag name at hash. If message is empty the
// tag name is used as the tag message so git never opens an editor.
func CreateTag(dir, name, message, hash string) error {
	if message == "" {
		message = name
	}
	cmd := exec.Command("git", "tag", "-a", name, "-m", message, hash)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func Ammend(dir, message string) error {
	cmd := exec.Command("git", "commit", "--amend", "-m", message)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// RewordCommit replaces the message of an arbitrary (not necessarily HEAD)
// commit in the current branch's history, via a scripted, non-interactive
// `git rebase -i`. Every commit after hash is replayed on top with its
// original tree unchanged - only hash's message and, as an unavoidable
// consequence, every descendant's hash, are rewritten.
//
// Rebasing needs the two spots where it would normally stop and hand
// control to an editor - the todo list and the commit message - driven
// non-interactively instead:
//   - GIT_SEQUENCE_EDITOR rewrites the first line of the todo list ("pick
//     <hash> ...") to "reword", since hash is always the oldest commit
//     being replayed (the rebase base is hash^) and so always lands first.
//   - GIT_EDITOR overwrites whatever commit-message file git hands it with
//     the message we already wrote to a temp file, sidestepping the need
//     to safely quote arbitrary commit-message text into a shell command.
func RewordCommit(dir, hash, message string) error {
	msgFile, err := os.CreateTemp("", "bubblegit-reword-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(msgFile.Name())
	if _, err := msgFile.WriteString(message); err != nil {
		msgFile.Close()
		return err
	}
	if err := msgFile.Close(); err != nil {
		return err
	}

	base := hash + "^"
	verify := exec.Command("git", "rev-parse", "--verify", "-q", base)
	verify.Dir = dir
	if err := verify.Run(); err != nil {
		// hash has no parent: it's the root commit.
		base = "--root"
	}

	cmd := exec.Command("git", "rebase", "-i", base)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_SEQUENCE_EDITOR=sed -i '1s/^pick /reword /'",
		"GIT_EDITOR=cp "+msgFile.Name(),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		abort := exec.Command("git", "rebase", "--abort")
		abort.Dir = dir
		_ = abort.Run()
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// StashEntry is one entry from `git stash list`.
type StashEntry struct {
	Ref     string // e.g. "stash@{0}"
	Date    string
	Message string
}

// Stashes lists the stash entries, most recent first.
func Stashes(dir string) ([]StashEntry, error) {
	// %gd must be requested without a --date flag: as soon as one is
	// present git renders the reflog selector as a date instead of the
	// numeric "stash@{N}" index. So the timestamp is pulled separately
	// via %ct and formatted here instead.
	format := strings.Join([]string{"%gd", "%ct", "%s"}, logFieldSep) + logRecordSep
	cmd := exec.Command("git", "stash", "list", "--pretty=format:"+format)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	if string(out) == "" {
		return nil, nil
	}
	var stashes []StashEntry
	for rec := range strings.SplitSeq(string(out), logRecordSep) {
		rec = strings.TrimPrefix(rec, "\n")
		if rec == "" {
			continue
		}
		f := strings.Split(rec, logFieldSep)
		if len(f) < 3 {
			continue
		}
		date := f[1]
		if ts, err := strconv.ParseInt(f[1], 10, 64); err == nil {
			date = time.Unix(ts, 0).Format("2006-01-02")
		}
		stashes = append(stashes, StashEntry{
			Ref:     f[0],
			Date:    date,
			Message: f[2],
		})
	}
	return stashes, nil
}

// StashPush stashes tracked changes (staged and unstaged), optionally
// under the given message.
func StashPush(dir, message string) error {
	args := []string{"stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// StashApply applies a stash entry to the working tree, leaving it in the
// stash list.
func StashApply(dir, ref string) error {
	cmd := exec.Command("git", "stash", "apply", ref)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// StashDrop removes a stash entry.
func StashDrop(dir, ref string) error {
	cmd := exec.Command("git", "stash", "drop", ref)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// StashBranch creates and checks out a new branch from the stash's
// original commit, applies the stash, and drops it from the list.
func StashBranch(dir, branch, ref string) error {
	cmd := exec.Command("git", "stash", "branch", branch, ref)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func StashPop(dir, ref string) error {
	cmd := exec.Command("git", "stash", "pop", ref)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// StashClear removes every stash entry.
func StashClear(dir string) error {
	cmd := exec.Command("git", "stash", "clear")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// StashShowDelta renders a stash entry's diff through delta. `git show` on
// a stash prints a combined "diff --cc" (it's a merge commit); `stash show
// -p` is the form that produces a normal unified diff.
func StashShow(dir, ref string) (string, error) {
	cmd := exec.Command("git", "stash", "show", "-p", "--color=always", ref)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func StashShowDelta(dir, ref string, sideBySide bool, width int) (string, error) {
	git := exec.Command("git", "stash", "show", "-p", "--no-color", ref)
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

func MergeFF(dir, branch string) error {
	return Merge(dir, branch, "ff")
}

// Merge merges branch into the current branch. mode is one of "ff"
// (fast-forward only), "merge" (create a merge commit even when a
// fast-forward is possible), or "squash" (squash all changes into a
// single commit).
func Merge(dir, branch, mode string) error {
	var cmd *exec.Cmd
	switch mode {
	case "ff":
		cmd = exec.Command("git", "merge", "--ff-only", branch)
	case "merge":
		cmd = exec.Command("git", "merge", "--no-ff", "--no-edit", branch)
	case "squash":
		cmd = exec.Command("git", "merge", "--squash", branch)
	default:
		return fmt.Errorf("unknown merge mode %q", mode)
	}
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	if mode == "squash" {
		commit := exec.Command("git", "commit", "-m", "Squash merge branch '"+branch+"'")
		commit.Dir = dir
		out, err = commit.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s", strings.TrimSpace(string(out)))
		}
	}
	return nil
}
