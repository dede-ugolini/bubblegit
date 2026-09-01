// Package git provides git operations
package git

import (
	"fmt"
	"os/exec"
	"strings"
)

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
