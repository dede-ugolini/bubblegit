package git

// FileStatus represents one line of `git status --porcelain` output.
type FileStatus struct {
	Index    byte
	Worktree byte
	Path     string

	OrigPath string
}

// Add stages the given path
func Add(dir, path string) error {
	_, err := runGitWrite(dir, "add", "--", path)
	return err
}

// AddAll stages all path
func AddAll(dir string) error {
	_, err := runGitWrite(dir, "add", "-A")
	return err
}

// Reset unstages the given path, leaving working tree changes intact.
func Reset(dir, path string) error {
	_, err := runGitWrite(dir, "reset", "--", path)
	return err
}

func ResetAll(dir string) error {
	_, err := runGitWrite(dir, "reset")
	return err
}

// Restore reverts the file to its HEAD state, discarding both staged and
// unstaged changes.
func Restore(dir, path string) error {
	_, err := runGitWrite(dir, "restore", "--staged", "--worktree", "--", path)
	return err
}

func RestoreUntracked(dir, path string) error {
	_, err := runGitWrite(dir, "clean", "-fd", path)
	return err
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
