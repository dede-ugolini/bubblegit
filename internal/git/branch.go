package git

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

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

// Push pushes branch to the origin remote, setting it as the upstream so a
// later pull needs no arguments. Errors (no origin remote, rejected
// non-fast-forward push, etc.) surface as the trimmed git stderr.
func Push(dir, branch string) error {
	cmd := exec.Command("git", "push", "-u", "origin", branch)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}
