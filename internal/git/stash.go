package git

import (
	"bytes"
	"fmt"
	"os/exec"
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
	cmd := exec.Command("git", "stash", "list", "--pretty=format:"+format)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	if string(out) == "" {
		return nil, nil
	}
	var stashes []StashEntry
	for rec := range strings.SplitSeq(string(out), logRecordSep) {
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
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// StashApply applies a stash entry to the working tree, leaving it in the
// stash list.
func StashApply(dir, ref string) error {
	cmd := exec.Command("git", "stash", "apply", ref)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// StashDrop removes a stash entry.
func StashDrop(dir, ref string) error {
	cmd := exec.Command("git", "stash", "drop", ref)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// StashBranch creates and checks out a new branch from the stash's
// original commit, applies the stash, and drops it from the list.
func StashBranch(dir, branch, ref string) error {
	cmd := exec.Command("git", "stash", "branch", branch, ref)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func StashPop(dir, ref string) error {
	cmd := exec.Command("git", "stash", "pop", ref)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// StashClear removes every stash entry.
func StashClear(dir string) error {
	cmd := exec.Command("git", "stash", "clear")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// StashShowDelta renders a stash entry's diff through delta. `git show` on
// a stash prints a combined "diff --cc" (it's a merge commit); `stash show
// -p` is the form that produces a normal unified diff.
func StashShow(dir, ref string) (string, error) {
	cmd := exec.Command("git", "stash", "show", "-p", "--color=always", ref)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func StashShowDelta(dir, ref string, sideBySide bool, width int) (string, error) {
	git := exec.Command("git", "stash", "show", "-p", "--no-color", ref)
	git.Dir = dir
	diff, err := git.Output()
	if err != nil {
		return "", err
	}
	delta := exec.Command(
		"delta",
		"--no-gitconfig",
		"--paging=never",
		"--line-numbers",
	)
	if sideBySide {
		delta.Args = append(delta.Args, "--side-by-side", fmt.Sprintf("--width=%d", width))
	}
	delta.Stdin = bytes.NewReader(diff)

	var output bytes.Buffer
	delta.Stdout = &output

	if err := delta.Run(); err != nil {
		return "", err
	}
	return output.String(), nil
}
