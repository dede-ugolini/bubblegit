package git

import "strings"

// CreateTag adds an annotated tag name at hash. If message is empty the
// tag name is used as the tag message so git never opens an editor.
func CreateTag(dir, name, message, hash string) error {
	if message == "" {
		message = name
	}
	_, err := runGitWrite(dir, "tag", "-a", name, "-m", message, hash)
	return err
}

// TagEntry is one tag as listed in the tags panel.
type TagEntry struct {
	Name string

	// Hash is the peeled commit the tag references: the commit an annotated
	// tag points through its tag object, or the commit itself for a
	// lightweight tag. It is what the diff panel shows for the selection.
	Hash string

	// Message is the tag message's subject line. Lightweight tags have no
	// message of their own, so git reports the tagged commit's subject.
	Message string
}

// Tags lists the repository's tags, newest first by creation date, each with
// the commit it points at and its message. It peels annotated tags to their
// commit the same way TagsByCommit does.
func Tags(dir string) ([]TagEntry, error) {
	format := strings.Join([]string{"%(refname:short)", "%(objectname)", "%(*objectname)", "%(contents:subject)"}, logFieldSep)
	out, err := runGit(dir, "for-each-ref", "refs/tags", "--sort=-creatordate", "--format="+format)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	var tags []TagEntry
	for line := range strings.SplitSeq(strings.TrimRight(out, "\n"), "\n") {
		f := strings.Split(line, logFieldSep)
		if len(f) < 4 {
			continue
		}
		hash := f[2]
		if hash == "" {
			hash = f[1]
		}
		tags = append(tags, TagEntry{Name: f[0], Hash: hash, Message: f[3]})
	}
	return tags, nil
}

// TagsByCommit maps each tagged commit hash to the tag names pointing at
// it. Annotated tags are peeled to the commit they reference; lightweight
// tags use their own object hash.
func TagsByCommit(dir string) (map[string][]string, error) {
	format := strings.Join([]string{"%(refname:short)", "%(objectname)", "%(*objectname)"}, logFieldSep)
	out, err := runGit(dir, "for-each-ref", "refs/tags", "--format="+format)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	tags := make(map[string][]string)
	for line := range strings.SplitSeq(strings.TrimRight(out, "\n"), "\n") {
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
