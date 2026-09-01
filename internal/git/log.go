package git

import (
	"fmt"
	"os/exec"
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
	cmd := exec.Command(
		"git",
		"log",
		rev,
		"--date=short",
		"--pretty=format:"+format,
		fmt.Sprintf("-n%d", limit),
	)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(err.Error(), "does not have any commits yet") ||
			strings.Contains(err.Error(), "unknown revision") {
			return nil, nil
		}
		return nil, err
	}
	if string(out) == "" {
		return nil, nil
	}
	var entries []LogEntry
	for rec := range strings.SplitSeq(string(out), logRecordSep) {
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
