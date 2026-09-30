package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// runGitDiff runs a diff-producing git command in dir. Exit status 1 is not
// treated as a failure: `git diff --no-index` uses it to mean "the files
// differ", which is the normal result for an untracked file rather than an
// error. Every other diff form below exits 0 whether or not anything
// differs, so nothing else is being excused here.
func runGitDiff(dir string, args ...string) (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return string(out), nil
		}
		return "", gitError(err, out)
	}
	return string(out), nil
}

// runDelta produces a diff with the given read-only git command and renders
// it through delta. The git command must ask for --no-color: delta does its
// own syntax highlighting, and pre-colored input arrives as literal escape
// sequences it can't interpret.
func runDelta(dir string, args []string, sideBySide bool, width int) (string, error) {
	diff, err := runGitDiff(dir, args...)
	if err != nil {
		return "", err
	}
	return renderDelta(diff, sideBySide, width)
}

// renderDelta pipes a unified diff through delta. Callers must not hold mu.
func renderDelta(diff string, sideBySide bool, width int) (string, error) {
	args := []string{"--no-gitconfig", "--paging=never", "--line-numbers"}
	if sideBySide {
		args = append(args, "--side-by-side", fmt.Sprintf("--width=%d", width))
	}
	cmd := exec.Command("delta", args...)
	cmd.Stdin = strings.NewReader(diff)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		// delta writes its own diagnostics (an unsupported option, a bad
		// config); surface them instead of a bare exit status.
		if msg := strings.TrimSpace(errOut.String()); msg != "" {
			return "", errors.New(msg)
		}
		return "", err
	}
	return out.String(), nil
}

func Diff(dir, path string) (string, error) {
	return runGit(dir, "diff", "--color=always", "--", path)
}

func DiffDelta(dir, path string, sideBySide bool, width int) (string, error) {
	return runDelta(dir, []string{"diff", "--no-color", "--", path}, sideBySide, width)
}

func DiffStaged(dir, path string) (string, error) {
	return runGit(dir, "diff", "--staged", "--color=always", "--", path)
}

func DiffDeltaStaged(dir, path string, sideBySide bool, width int) (string, error) {
	return runDelta(dir, []string{"diff", "--staged", "--no-color", "--", path}, sideBySide, width)
}

func DiffUntracked(dir, path string) (string, error) {
	return runGitDiff(dir, "diff", "--no-index", "--color=always", "/dev/null", "--", path)
}

func DiffDeltaUntracked(dir, path string, sideBySide bool, width int) (string, error) {
	return runDelta(dir, []string{"diff", "--no-index", "--no-color", "/dev/null", "--", path}, sideBySide, width)
}

func DiffBranch(dir string) (string, error) {
	return runGit(dir, "diff", "--color=always", "--stat", "--patch")
}

func DiffBranchDelta(dir string, sideBySide bool, width int) (string, error) {
	return runDelta(dir, []string{"diff", "--no-color", "--stat", "--patch"}, sideBySide, width)
}
