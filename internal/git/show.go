package git

import (
	"bytes"
	"fmt"
	"os/exec"
)

func Show(dir, hash string) (string, error) {
	cmd := exec.Command("git", "show", "--color=always", "--stat", "--patch", hash)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(""), err
	}
	return string(out), nil
}

func ShowDelta(dir, path string, sideBySide bool, width int) (string, error) {
	git := exec.Command("git", "show", "--color=always", "--stat", "--patch", path)
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
