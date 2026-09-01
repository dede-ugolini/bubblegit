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
