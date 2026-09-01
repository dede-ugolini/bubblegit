package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// SetRemote points the "origin" remote at url, adding it if this repo has
// none yet or repointing the existing one otherwise - so it works the same
// the first time a repo gets a remote and to correct one later.
func SetRemote(dir, url string) error {
	list := exec.Command("git", "remote")
	list.Dir = dir
	out, err := list.Output()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}

	args := []string{"remote", "add", "origin", url}
	for _, name := range strings.Fields(string(out)) {
		if name == "origin" {
			args = []string{"remote", "set-url", "origin", url}
			break
		}
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}
