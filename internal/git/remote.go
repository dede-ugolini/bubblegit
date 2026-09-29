package git

import "strings"

// SetRemote points the "origin" remote at url, adding it if this repo has
// none yet or repointing the existing one otherwise - so it works the same
// the first time a repo gets a remote and to correct one later.
func SetRemote(dir, url string) error {
	mu.Lock()
	defer mu.Unlock()

	out, err := execGit(dir, nil, "remote")
	if err != nil {
		return err
	}
	args := []string{"remote", "add", "origin", url}
	for _, name := range strings.Fields(out) {
		if name == "origin" {
			args = []string{"remote", "set-url", "origin", url}
			break
		}
	}
	_, err = execGit(dir, nil, args...)
	return err
}
