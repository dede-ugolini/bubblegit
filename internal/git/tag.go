package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// CreateTag adds an annotated tag name at hash. If message is empty the
// tag name is used as the tag message so git never opens an editor.
func CreateTag(dir, name, message, hash string) error {
	if message == "" {
		message = name
	}
	cmd := exec.Command("git", "tag", "-a", name, "-m", message, hash)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// TagsByCommit maps each tagged commit hash to the tag names pointing at
// it. Annotated tags are peeled to the commit they reference; lightweight
// tags use their own object hash.
func TagsByCommit(dir string) (map[string][]string, error) {
	format := strings.Join([]string{"%(refname:short)", "%(objectname)", "%(*objectname)"}, logFieldSep)
	cmd := exec.Command("git", "for-each-ref", "refs/tags", "--format="+format)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	if string(out) == "" {
		return nil, nil
	}
	tags := make(map[string][]string)
	for line := range strings.SplitSeq(strings.TrimRight(string(out), "\n"), "\n") {
		f := strings.Split(line, logFieldSep)
		if len(f) < 3 {
			continue
		}
		hash := f[2]
		if hash == "" {
			hash = f[1]
		}
		tags[hash] = append(tags[hash], f[0])
	}
	return tags, nil
}
