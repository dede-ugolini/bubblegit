package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func Diff(dir, path string) (string, error) {
	cmd := exec.Command("git", "diff", "--color=always", "--", path)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(""), err
	}
	return string(out), nil
}

func DiffDelta(dir, path string, sideBySide bool, width int) (string, error) {
	git := exec.Command("git", "diff", "--no-color", "--", path)
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

func DiffStaged(dir, path string) (string, error) {
	cmd := exec.Command("git", "diff", "--staged", "--color=always", "--", path)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(""), err
	}
	return string(out), nil
}

func DiffDeltaStaged(dir, path string, sideBySide bool, width int) (string, error) {
	git := exec.Command("git", "diff", "--staged", "--no-color", "--", path)
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

func DiffUntracked(dir, path string) (string, error) {
	cmd := exec.Command("git", "diff", "--no-index", "--color=always", "/dev/null", "--", path)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return string(out), nil
}

func DiffDeltaUntracked(dir, path string, sideBySide bool, width int) (string, error) {
	git := exec.Command("git", "diff", "--no-index", "--no-color", "/dev/null", "--", path)
	git.Dir = dir
	diff, _ := git.Output()
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

func DiffBranch(dir string) (string, error) {
	cmd := exec.Command("git", "diff", "--color=always", "--stat", "--patch")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func DiffBranchDelta(dir string, sideBySide bool, width int) (string, error) {
	git := exec.Command("git", "diff", "--color=always", "--stat", "--patch")
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
