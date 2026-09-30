package git

func Show(dir, hash string) (string, error) {
	return runGit(dir, "show", "--color=always", "--stat", "--patch", hash)
}

func ShowDelta(dir, hash string, sideBySide bool, width int) (string, error) {
	return runDelta(dir, []string{"show", "--no-color", "--stat", "--patch", hash}, sideBySide, width)
}
