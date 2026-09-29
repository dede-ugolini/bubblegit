package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type BranchInfo struct {
	Name    string
	Current bool

	// Upstream is the remote-tracking branch this branch tracks, e.g.
	// "origin/main"; empty if it has none.
	Upstream string

	// Ahead counts local commits not on the upstream; Behind counts the
	// reverse. Both are zero when in sync or when there is no upstream.
	Ahead  int
	Behind int

	// Gone reports an upstream that no longer exists on the remote.
	Gone bool

	// CommitTime is the committer time of the branch tip; the zero value
	// means the branch has no commits yet (unborn).
	CommitTime time.Time
}

var (
	aheadRe  = regexp.MustCompile(`ahead (\d+)`)
	behindRe = regexp.MustCompile(`behind (\d+)`)
)

// Branches lists local branches, current branch first, each with the
// ahead/behind counts relative to its upstream as reported by git.
func Branches(dir string) ([]BranchInfo, error) {
	mu.RLock()
	defer mu.RUnlock()
	cmd := exec.Command(
		"git",
		"branch",
		"--format=%(HEAD)"+logFieldSep+"%(refname)"+logFieldSep+"%(refname:short)"+logFieldSep+"%(upstream:short)"+logFieldSep+"%(upstream:track)"+logFieldSep+"%(committerdate:unix)",
	)
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
		f := strings.SplitN(line, logFieldSep, 6)
		if len(f) != 6 {
			continue
		}
		// In detached-HEAD state, git lists a synthetic
		// "(HEAD detached at ...)" row whose refname isn't a real ref
		// under refs/heads/ — skip it, it's not a checkout/delete target.
		if !strings.HasPrefix(f[1], "refs/heads/") {
			continue
		}
		b := BranchInfo{Name: f[2], Current: f[0] == "*", Upstream: f[3]}
		if ts, err := strconv.ParseInt(f[5], 10, 64); err == nil {
			b.CommitTime = time.Unix(ts, 0)
		}
		if track := f[4]; track != "" {
			b.Gone = strings.Contains(track, "gone")
			if m := aheadRe.FindStringSubmatch(track); m != nil {
				b.Ahead, _ = strconv.Atoi(m[1])
			}
			if m := behindRe.FindStringSubmatch(track); m != nil {
				b.Behind, _ = strconv.Atoi(m[1])
			}
		}
		branches = append(branches, b)
	}
	sort.SliceStable(branches, func(i, j int) bool {
		return branches[i].Current && !branches[j].Current
	})
	return branches, nil
}

// Checkout switches the working tree to the given branch.
func Checkout(dir, branch string) error {
	mu.Lock()
	defer mu.Unlock()
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
	mu.Lock()
	defer mu.Unlock()
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
	mu.Lock()
	defer mu.Unlock()
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
	mu.Lock()
	defer mu.Unlock()
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
	mu.Lock()
	defer mu.Unlock()
	cmd := exec.Command("git", "push", "-u", "origin", branch)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}
