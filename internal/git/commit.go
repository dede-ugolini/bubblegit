package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Commit(dir, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func Ammend(dir, message string) error {
	cmd := exec.Command("git", "commit", "--amend", "-m", message)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// DropLastCommit removes the most recent commit (HEAD) from the current
// branch, discarding its changes. It refuses to drop the only (root) commit.
func DropLastCommit(dir string) error {
	verify := exec.Command("git", "rev-parse", "--verify", "-q", "HEAD^")
	verify.Dir = dir
	if err := verify.Run(); err != nil {
		return fmt.Errorf("cannot drop the only commit")
	}
	cmd := exec.Command("git", "reset", "--hard", "HEAD^")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// RewordCommit replaces the message of an arbitrary (not necessarily HEAD)
// commit in the current branch's history, via a scripted, non-interactive
// `git rebase -i`. Every commit after hash is replayed on top with its
// original tree unchanged - only hash's message and, as an unavoidable
// consequence, every descendant's hash, are rewritten.
//
// Rebasing needs the two spots where it would normally stop and hand
// control to an editor - the todo list and the commit message - driven
// non-interactively instead:
//   - GIT_SEQUENCE_EDITOR rewrites the first line of the todo list ("pick
//     <hash> ...") to "reword", since hash is always the oldest commit
//     being replayed (the rebase base is hash^) and so always lands first.
//   - GIT_EDITOR overwrites whatever commit-message file git hands it with
//     the message we already wrote to a temp file, sidestepping the need
//     to safely quote arbitrary commit-message text into a shell command.
func RewordCommit(dir, hash, message string) error {
	msgFile, err := os.CreateTemp("", "bubblegit-reword-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(msgFile.Name())
	if _, err := msgFile.WriteString(message); err != nil {
		msgFile.Close()
		return err
	}
	if err := msgFile.Close(); err != nil {
		return err
	}

	base := hash + "^"
	verify := exec.Command("git", "rev-parse", "--verify", "-q", base)
	verify.Dir = dir
	if err := verify.Run(); err != nil {
		// hash has no parent: it's the root commit.
		base = "--root"
	}

	cmd := exec.Command("git", "rebase", "-i", base)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_SEQUENCE_EDITOR=sed -i '1s/^pick /reword /'",
		"GIT_EDITOR=cp "+msgFile.Name(),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		abort := exec.Command("git", "rebase", "--abort")
		abort.Dir = dir
		_ = abort.Run()
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// SquashCommits folds `count` consecutive commits, starting at the oldest
// commit oldestHash, into a single commit carrying message. Like
// RewordCommit, it drives a non-interactive `git rebase -i` scripted via
// GIT_SEQUENCE_EDITOR/GIT_EDITOR.
//
// oldestHash is always first in the rebase todo list (base is oldestHash^),
// so the first `count` lines of the todo are exactly the range to fold:
// line 1 stays "pick" (the commit the rest squash onto) and lines 2..count
// are rewritten from "pick" to "squash". squash (not fixup) is used because
// it still pauses for a combined commit-message edit the same way reword
// does, so GIT_EDITOR=cp overwrites it with our own message - identical
// mechanism to RewordCommit, no extra step needed.
func SquashCommits(dir, oldestHash string, count int, message string) error {
	if count < 2 {
		return fmt.Errorf("need at least 2 commits to squash")
	}
	msgFile, err := os.CreateTemp("", "bubblegit-squash-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(msgFile.Name())
	if _, err := msgFile.WriteString(message); err != nil {
		msgFile.Close()
		return err
	}
	if err := msgFile.Close(); err != nil {
		return err
	}

	base := oldestHash + "^"
	verify := exec.Command("git", "rev-parse", "--verify", "-q", base)
	verify.Dir = dir
	if err := verify.Run(); err != nil {
		// oldestHash has no parent: it's the root commit.
		base = "--root"
	}

	cmd := exec.Command("git", "rebase", "-i", base)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_SEQUENCE_EDITOR=sed -i '2,%ds/^pick /squash /'", count),
		"GIT_EDITOR=cp "+msgFile.Name(),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		abort := exec.Command("git", "rebase", "--abort")
		abort.Dir = dir
		_ = abort.Run()
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}
