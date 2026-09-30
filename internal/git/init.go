package git

import "strings"

// IsRepo reports whether dir is inside a git work tree. Subdirectories of a
// repository count; bare and otherwise uninitialized directories don't.
func IsRepo(dir string) bool {
	out, err := runGit(dir, "rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

// Init creates a new repository in dir. A non-empty branch names the initial
// branch via -b; an empty one leaves git's default (init.defaultBranch or
// master) in charge.
func Init(dir, branch string) error {
	args := []string{"init"}
	if branch != "" {
		args = append(args, "-b", branch)
	}
	_, err := runGitWrite(dir, args...)
	return err
}
