// Package git provides git operations
package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// field/record separators unlikely to appear in commit metadata.
const (
	logFieldSep  = "\x1f"
	logRecordSep = "\x1e"
)

// mu serializes git subprocesses. The TUI dispatches commands from
// concurrent goroutines (batched refreshes, per-keypress diffs, and
// mutating actions can all overlap); git itself only tolerates concurrent
// reads, and two overlapping index/ref writers collide on index.lock or
// produce inconsistent state. So writes take an exclusive lock while reads
// share it, keeping the batched parallel refresh fast.
var mu sync.RWMutex

// execGit runs git with args in dir and returns its combined stdout+stderr.
// A failure becomes an error carrying git's own message, which is what the
// UI shows the user. env, when set, is appended to the inherited
// environment.
//
// Callers must already hold mu, shared or exclusive - see runGit and
// runGitWrite. Operations spanning more than one git invocation (a merge
// plus its commit, a rebase plus its abort) hold the lock across the whole
// sequence and call execGit directly.
func execGit(dir string, env []string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", gitError(err, out)
	}
	return string(out), nil
}

// gitError prefers git's own message over the bare exit status, so a
// failure reads "fatal: not a git repository" rather than "exit status
// 128". Note that exec.ExitError.Error() is only ever the status - the
// message lives in the command's output - so this has to be built from
// `out`, not from `err`. When git printed nothing at all, the exec error is
// all there is, and beats an empty message.
func gitError(err error, out []byte) error {
	if msg := strings.TrimSpace(string(out)); msg != "" {
		return errors.New(msg)
	}
	return err
}

// runGit runs a read-only git command in dir. Reads share the lock, so the
// TUI's batched refresh still fans out in parallel.
func runGit(dir string, args ...string) (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	return execGit(dir, nil, args...)
}

// runGitWrite runs a git command that mutates the repository. Writes take
// the lock exclusively; see mu.
func runGitWrite(dir string, args ...string) (string, error) {
	mu.Lock()
	defer mu.Unlock()
	return execGit(dir, nil, args...)
}

// Status returns the working tree status
func Status(dir string) ([]FileStatus, error) {
	out, err := runGit(dir, "status", "--porcelain", "-z", "--untracked-files=all")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return parseStatus(out)
}

func parseStatus(out string) ([]FileStatus, error) {
	var files []FileStatus
	parts := strings.Split(strings.TrimRight(out, "\x00"), "\x00")

	for i := 0; i < len(parts); i++ {
		entry := parts[i]
		if len(entry) < 3 {
			continue
		}
		fs := FileStatus{
			Index:    entry[0],
			Worktree: entry[1],
			Path:     entry[3:],
		}
		// Renames/copies have an extra NUL-separated original path.
		if fs.Index == 'R' || fs.Index == 'C' {
			i++
			if i < len(parts) {
				fs.OrigPath = parts[i]
			}
		}
		files = append(files, fs)
	}
	return files, nil
}

// Merge merges branch into the current branch. mode is one of "ff"
// (fast-forward only), "merge" (create a merge commit even when a
// fast-forward is possible), or "squash" (squash all changes into a
// single commit).
func Merge(dir, branch, mode string) error {
	mu.Lock()
	defer mu.Unlock()

	var args []string
	switch mode {
	case "ff":
		args = []string{"merge", "--ff-only", branch}
	case "merge":
		args = []string{"merge", "--no-ff", "--no-edit", branch}
	case "squash":
		args = []string{"merge", "--squash", branch}
	default:
		return fmt.Errorf("unknown merge mode %q", mode)
	}
	if _, err := execGit(dir, nil, args...); err != nil {
		return err
	}
	if mode == "squash" {
		// --squash stages the merge but records no commit, so the folded
		// changes still need one.
		_, err := execGit(dir, nil, "commit", "-m", "Squash merge branch '"+branch+"'")
		return err
	}
	return nil
}
