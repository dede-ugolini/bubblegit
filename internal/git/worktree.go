package git

import (
	"strings"
)

type WorktreeEntry struct {
	path, hash, branch string
}

func Worktrees(dir string) ([]WorktreeEntry, error) {
	out, err := runGit(dir, "worktree", "list")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	var worktrees []WorktreeEntry
	for line := range strings.SplitSeq(out, "\n") {
		fields := strings.Fields(line)
		path := fields[0]
		hash := fields[1]
		branch := fields[2]
		branch = strings.TrimPrefix(branch, "[")
		branch = strings.TrimSuffix(branch, "]")
		worktrees = append(worktrees, WorktreeEntry{path: path, hash: hash, branch: branch})
	}
	return worktrees, nil
}
