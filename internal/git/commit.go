package git

import (
	"fmt"
	"os"
)

func Commit(dir, message string) error {
	_, err := runGitWrite(dir, "commit", "-m", message)
	return err
}

func Ammend(dir, message string) error {
	_, err := runGitWrite(dir, "commit", "--amend", "-m", message)
	return err
}

// DropLastCommit removes the most recent commit (HEAD) from the current
// branch, discarding its changes. It refuses to drop the only (root) commit.
func DropLastCommit(dir string) error {
	mu.Lock()
	defer mu.Unlock()
	if _, err := execGit(dir, nil, "rev-parse", "--verify", "-q", "HEAD^"); err != nil {
		return fmt.Errorf("cannot drop the only commit")
	}
	_, err := execGit(dir, nil, "reset", "--hard", "HEAD^")
	return err
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
	msgFile, err := writeMsgFile("reword", message)
	if err != nil {
		return err
	}
	defer os.Remove(msgFile)

	base := hash + "^"
	if _, err := runGit(dir, "rev-parse", "--verify", "-q", base); err != nil {
		// hash has no parent: it's the root commit.
		base = "--root"
	}
	return rebaseRewrite(dir, msgFile, "GIT_SEQUENCE_EDITOR=sed -i '1s/^pick /reword /'",
		"rebase", "-i", base)
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
	msgFile, err := writeMsgFile("squash", message)
	if err != nil {
		return err
	}
	defer os.Remove(msgFile)

	base := oldestHash + "^"
	if _, err := runGit(dir, "rev-parse", "--verify", "-q", base); err != nil {
		// oldestHash has no parent: it's the root commit.
		base = "--root"
	}
	editor := fmt.Sprintf("GIT_SEQUENCE_EDITOR=sed -i '2,%ds/^pick /squash /'", count)
	return rebaseRewrite(dir, msgFile, editor, "rebase", "-i", base)
}

// writeMsgFile writes message to a temp file for GIT_EDITOR to copy over
// git's commit-message file, and returns its path. RewordCommit and
// SquashCommits are the only callers; the caller owns removing the file.
func writeMsgFile(kind, message string) (string, error) {
	f, err := os.CreateTemp("", "bubblegit-"+kind+"-*.txt")
	if err != nil {
		return "", err
	}
	if _, err := f.WriteString(message); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

// rebaseRewrite runs the scripted `git rebase -i` shared by RewordCommit and
// SquashCommits: sequenceEditor rewrites the todo list, and the message in
// msgFile is what GIT_EDITOR copies into place. The rebase and its cleanup
// share one exclusive lock, so nothing can interleave between a failure and
// the abort that recovers from it.
func rebaseRewrite(dir, msgFile, sequenceEditor string, args ...string) error {
	mu.Lock()
	defer mu.Unlock()
	_, err := execGit(dir, []string{
		sequenceEditor,
		"GIT_EDITOR=cp " + msgFile,
	}, args...)
	if err == nil {
		return nil
	}
	// Don't leave a half-finished rebase for the user to discover and clean
	// up by hand.
	_, _ = execGit(dir, nil, "rebase", "--abort")
	return err
}
