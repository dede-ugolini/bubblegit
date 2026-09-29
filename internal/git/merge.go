package git

import "fmt"

// MergeMode selects the strategy Merge applies.
type MergeMode int

const (
	// MergeFF fast-forwards only, and fails if a fast-forward isn't
	// possible.
	MergeFF MergeMode = iota

	// MergeCommit always records a merge commit, even when a
	// fast-forward would have been possible.
	MergeCommit

	// MergeSquash folds branch's changes into the index and working tree
	// as staged changes without recording a commit.
	MergeSquash
)

// MergeModes lists every mode, in the order a UI should offer them.
var MergeModes = []MergeMode{MergeFF, MergeCommit, MergeSquash}

// String names the mode for display.
func (m MergeMode) String() string {
	switch m {
	case MergeFF:
		return "fast-forward"
	case MergeCommit:
		return "merge commit"
	case MergeSquash:
		return "squash commit"
	default:
		return "unknown"
	}
}

// args returns the git command implementing the mode, or nil for a value
// that isn't one of the modes above - which Merge reports as an error
// rather than shelling out to git.
func (m MergeMode) args(branch string) []string {
	switch m {
	case MergeFF:
		return []string{"merge", "--ff-only", branch}
	case MergeCommit:
		return []string{"merge", "--no-ff", "--no-edit", branch}
	case MergeSquash:
		return []string{"merge", "--squash", branch}
	}
	return nil
}

// Merge merges branch into the current branch using mode.
func Merge(dir, branch string, mode MergeMode) error {
	mu.Lock()
	defer mu.Unlock()

	args := mode.args(branch)
	if args == nil {
		return fmt.Errorf("unknown merge mode %d", int(mode))
	}
	if _, err := execGit(dir, nil, args...); err != nil {
		return err
	}
	if mode == MergeSquash {
		// --squash stages the merge but records no commit, so the folded
		// changes still need one.
		_, err := execGit(dir, nil, "commit", "-m", "Squash merge branch '"+branch+"'")
		return err
	}
	return nil
}
