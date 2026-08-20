// Package git provides git operations
package git

import (
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

// Checkout switches the working tree to the given branch.
func Checkout(dir, branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = dir
	return cmd.Run()
}

// CreateBranch create and checks out a new branch off the current HEAD.
func CreateBranch(dir, branch string) error {
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = dir
	return cmd.Run()
}

// DeleteBranch removes a local branch. It refuses (like plain `git branch
// -d`) if the branch has commits not merged elsewhere.
func DeleteBranch(dir, branch string) error {
	cmd := exec.Command("git", "branch", "-d", branch)
	cmd.Dir = dir
	return cmd.Run()
}
