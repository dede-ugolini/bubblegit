package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// IsRepo reports whether dir is inside a git work tree. Subdirectories of a
// repository count; bare and otherwise uninitialized directories don't.
func IsRepo(dir string) bool {
	mu.RLock()
	defer mu.RUnlock()
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// Init creates a new repository in dir. A non-empty branch names the initial
// branch via -b; an empty one leaves git's default (init.defaultBranch or
// master) in charge.
func Init(dir, branch string) error {
	mu.Lock()
	defer mu.Unlock()
	args := []string{"init"}
	if branch != "" {
		args = append(args, "-b", branch)
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}
