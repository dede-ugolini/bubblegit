package git

import (
	"strings"
)

// WorktreeEntry is one working tree linked to the repository, as reported by
// `git worktree list`.
type WorktreeEntry struct {
	Path      string
	Hash      string
	ShortHash string

	// Branch is the short name of the branch checked out in the worktree, or
	// "(detached)" / "(bare)" when it is on no branch.
	Branch string
}

// Worktrees lists the repository's working trees. It parses the porcelain
// format, not the default table: the table is space-aligned, so a path
// containing a space - or a branchless worktree on a bare repository - would
// misparse, and its trailing newline would yield an empty record. Porcelain
// is one attribute per line, with a blank line between entries.
func Worktrees(dir string) ([]WorktreeEntry, error) {
	out, err := runGit(dir, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}

	var (
		worktrees []WorktreeEntry
		cur       WorktreeEntry
	)
	flush := func() {
		if cur.Path == "" {
			return
		}
		worktrees = append(worktrees, cur)
		cur = WorktreeEntry{}
	}
	for line := range strings.SplitSeq(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			// Blank lines separate entries, but flush here too so a missing
			// separator can't merge two worktrees into one.
			flush()
			cur.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			cur.Hash = strings.TrimPrefix(line, "HEAD ")
			cur.ShortHash = cur.Hash
			if len(cur.ShortHash) > 7 {
				cur.ShortHash = cur.ShortHash[:7]
			}
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			cur.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "detached":
			cur.Branch = "(detached)"
		case line == "bare":
			cur.Branch = "(bare)"
		case line == "":
			flush()
		}
	}
	flush()
	return worktrees, nil
}

// DeleteWorktree removes the linked worktree at path.
func DeleteWorktree(dir, path string) error {
	_, err := runGitWrite(dir, "worktree", "remove", path)
	return err
}

// AddWorktree links a new working tree at path. An empty branch lets git
// name the branch from path's basename; create true makes a new branch with
// -b, false checks out a branch that must already exist. Like the other
// worktree operations it mutates the repository, so it takes the write lock.
func AddWorktree(dir, path, branch string, create bool) error {
	args := []string{"worktree", "add"}
	if branch != "" && create {
		args = append(args, "-b", branch)
	}
	args = append(args, path)
	if branch != "" && !create {
		args = append(args, branch)
	}
	_, err := runGitWrite(dir, args...)
	return err
}
