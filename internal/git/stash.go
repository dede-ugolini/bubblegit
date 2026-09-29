package git

import (
	"strconv"
	"strings"
	"time"
)

// StashEntry is one entry from `git stash list`.
type StashEntry struct {
	Ref     string // e.g. "stash@{0}"
	Date    string
	Message string
}

// Stashes lists the stash entries, most recent first.
func Stashes(dir string) ([]StashEntry, error) {
	// %gd must be requested without a --date flag: as soon as one is
	// present git renders the reflog selector as a date instead of the
	// numeric "stash@{N}" index. So the timestamp is pulled separately
	// via %ct and formatted here instead.
	format := strings.Join([]string{"%gd", "%ct", "%s"}, logFieldSep) + logRecordSep
	out, err := runGit(dir, "stash", "list", "--pretty=format:"+format)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	var stashes []StashEntry
	for rec := range strings.SplitSeq(out, logRecordSep) {
		rec = strings.TrimPrefix(rec, "\n")
		if rec == "" {
			continue
		}
		f := strings.Split(rec, logFieldSep)
		if len(f) < 3 {
			continue
		}
		date := f[1]
		if ts, err := strconv.ParseInt(f[1], 10, 64); err == nil {
			date = time.Unix(ts, 0).Format("2006-01-02")
		}
		stashes = append(stashes, StashEntry{
			Ref:     f[0],
			Date:    date,
			Message: f[2],
		})
	}
	return stashes, nil
}

// StashPush stashes tracked changes (staged and unstaged), optionally
// under the given message.
func StashPush(dir, message string) error {
	args := []string{"stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}
	_, err := runGitWrite(dir, args...)
	return err
}

// StashApply applies a stash entry to the working tree, leaving it in the
// stash list.
func StashApply(dir, ref string) error {
	_, err := runGitWrite(dir, "stash", "apply", ref)
	return err
}

// StashDrop removes a stash entry.
func StashDrop(dir, ref string) error {
	_, err := runGitWrite(dir, "stash", "drop", ref)
	return err
}

// StashBranch creates and checks out a new branch from the stash's
// original commit, applies the stash, and drops it from the list.
func StashBranch(dir, branch, ref string) error {
	_, err := runGitWrite(dir, "stash", "branch", branch, ref)
	return err
}

func StashPop(dir, ref string) error {
	_, err := runGitWrite(dir, "stash", "pop", ref)
	return err
}

// StashClear removes every stash entry.
func StashClear(dir string) error {
	_, err := runGitWrite(dir, "stash", "clear")
	return err
}

// StashShow returns a stash entry's diff. `git show` on a stash prints a
// combined "diff --cc" (it's a merge commit); `stash show -p` is the form
// that produces a normal unified diff.
func StashShow(dir, ref string) (string, error) {
	return runGit(dir, "stash", "show", "-p", "--color=always", ref)
}

func StashShowDelta(dir, ref string, sideBySide bool, width int) (string, error) {
	return runDelta(dir, []string{"stash", "show", "-p", "--no-color", ref}, sideBySide, width)
}
