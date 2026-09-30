package git

import (
	"fmt"
	"strings"
)

// LogEntry is one commit as reported by `git log`.
type LogEntry struct {
	Hash      string
	ShortHash string
	Author    string
	Date      string
	Subject   string
	Body      string
}

func Log(dir, rev string, limit int) ([]LogEntry, error) {
	format := strings.Join(
		[]string{"%H", "%h", "%an", "%ad", "%s", "%b"}, logFieldSep,
	) + logRecordSep
	out, err := runGit(
		dir, "log", rev, "--date=short",
		"--pretty=format:"+format, fmt.Sprintf("-n%d", limit),
	)
	if err != nil {
		// A branch with no commits yet has no HEAD to log. That is an
		// empty log panel, not a failure, so it is reported as no entries.
		// runGit folds git's own message into err, which is what has to be
		// matched here - err.Error() on its own is only "exit status 128".
		if strings.Contains(err.Error(), "does not have any commits yet") ||
			strings.Contains(err.Error(), "unknown revision") {
			return nil, nil
		}
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	var entries []LogEntry
	for rec := range strings.SplitSeq(out, logRecordSep) {
		rec = strings.TrimPrefix(rec, "\n")
		if rec == "" {
			continue
		}
		f := strings.Split(rec, logFieldSep)
		if len(f) < 6 {
			continue
		}
		entries = append(entries, LogEntry{
			Hash:      f[0],
			ShortHash: f[1],
			Author:    f[2],
			Date:      f[3],
			Subject:   f[4],
			Body:      strings.TrimRight(f[5], "\n"),
		})
	}
	return entries, nil
}

// AheadHashes returns the full hashes of the commits reachable from HEAD but
// not from its upstream - what a push would publish. The set is empty for a
// branch with no upstream, a missing upstream, or nothing ahead.
func AheadHashes(dir string) (map[string]bool, error) {
	out, err := runGit(dir, "log", "@{upstream}..HEAD", "--pretty=%H")
	if err != nil {
		// No upstream to compare against (detached HEAD, untracked branch,
		// gone upstream) means nothing is ahead.
		return nil, nil
	}
	hashes := make(map[string]bool)
	for _, h := range strings.Fields(out) {
		hashes[h] = true
	}
	return hashes, nil
}
