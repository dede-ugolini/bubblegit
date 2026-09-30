package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// unmergedCodes are the two-letter `git status --porcelain` codes for an
// unmerged path. Unlike every other code, both letters are always taken from
// {D, A, U} and the pair describes how the two sides disagree, so an entry is
// unmerged exactly when its two bytes spell one of these.
var unmergedCodes = map[[2]byte]bool{
	{'D', 'D'}: true, // deleted by both
	{'A', 'U'}: true, // added by us, deleted by them
	{'U', 'D'}: true, // deleted by us, added by them
	{'U', 'A'}: true, // added by them
	{'D', 'U'}: true, // deleted by us
	{'A', 'A'}: true, // added by both
	{'U', 'U'}: true, // modified by both
}

// Conflicted reports whether this entry is unmerged: the two sides of a merge
// disagreed and the path has no single content until the conflict is resolved.
//
// This has to be checked before Staged, because an unmerged entry reports a
// non-space index byte and so Staged claims it - which is true in the sense
// that `git commit` will refuse until it is resolved, but misleading for a
// status column that means "changes ready to commit".
func (f FileStatus) Conflicted() bool {
	return unmergedCodes[[2]byte{f.Index, f.Worktree}]
}

// Resolution is what to keep of one conflicted hunk.
type Resolution int

const (
	// ResolveOurs keeps the side from HEAD.
	ResolveOurs Resolution = iota
	// ResolveTheirs keeps the side from the branch being merged in.
	ResolveTheirs
	// ResolveBoth keeps ours and then theirs, in that order. Text at the
	// seam is often wrong, so it is offered as a starting point rather than
	// a correct answer.
	ResolveBoth
)

// Resolutions lists every resolution in the order a UI should offer them.
var Resolutions = []Resolution{ResolveOurs, ResolveTheirs, ResolveBoth}

// String names the resolution for display.
func (r Resolution) String() string {
	switch r {
	case ResolveOurs:
		return "ours"
	case ResolveTheirs:
		return "theirs"
	case ResolveBoth:
		return "both"
	default:
		return "unknown"
	}
}

// Hunk is one conflicted region of a file.
type Hunk struct {
	// Ours and Theirs are the two sides, split into lines with any trailing
	// newline removed. Ours is the HEAD side and Theirs the merged-in side,
	// matching git's own marker labels.
	Ours   []string
	Theirs []string

	// Base is the common ancestor, and HasBase reports whether there was one.
	// MergeConflict always produces it, because it asks git for diff3 markers;
	// it is empty for an add/add conflict, where the file has no ancestor.
	Base    []string
	HasBase bool
}

// ConflictFile is a file's conflicted regions, split apart. The two lists
// interleave: Blocks[i] is the text preceding Hunks[i], and Blocks always has
// exactly one more entry than Hunks, so the text after the last conflict is
// Blocks[len(Hunks)].
type ConflictFile struct {
	Blocks [][]string
	Hunks  []Hunk

	// trailingNewline records whether the file ended with a newline, so
	// rendering reproduces the original file's line endings rather than
	// dropping or inventing one.
	trailingNewline bool
}

// markerWidth is how many repeated characters make up a conflict marker. git
// always uses seven, and widens it when the surrounding text would otherwise
// collide; a run of fewer than seven is never a marker.
const markerWidth = 7

// splitMarker reports whether line begins with a run of at least markerWidth
// copies of c, and returns whatever follows the run.
//
// The `=======` separator is the one marker git writes with nothing after it,
// so its remainder must be blank. Accepting trailing junk there would let an
// ordinary line of equals signs in the middle of a hunk read as a separator.
func splitMarker(line string, c byte) (rest string, marker bool) {
	n := 0
	for n < len(line) && line[n] == c {
		n++
	}
	if n < markerWidth {
		return "", false
	}
	rest = line[n:]
	if c == '=' && strings.TrimSpace(rest) != "" {
		return "", false
	}
	return rest, true
}

// parseSide is which part of a hunk subsequent lines belong to.
type parseSide int

const (
	sideOurs parseSide = iota
	sideBase
	sideTheirs
)

// ParseConflict splits conflict markers into hunks.
//
// Markers are counted by depth rather than matched line by line, so that a
// malformed file - one whose marker runs do not balance - is reported as an
// error instead of being resolved into a guess. Writing back a mis-parsed file
// would silently discard one side of someone's merge, which is the one failure
// mode a conflict resolver must not have. Nothing produces nested markers, but
// handling them costs a slice index.
//
// MergeConflict is what feeds this in practice, and it feeds it only text git
// generated itself, so the error path is insurance rather than routine.
func ParseConflict(content string) (*ConflictFile, error) {
	lines, trailing := splitLines(content)
	f := &ConflictFile{trailingNewline: trailing}

	// Blocks starts with one entry and grows by one for every conflict, so
	// Blocks[i] is always the run of plain lines just before Hunks[i] - which
	// is what makes the interleaving hold with zero or many conflicts.
	f.Blocks = [][]string{nil}

	var (
		stack []Hunk
		sides []parseSide
	)

	for i, line := range lines {
		if _, marker := splitMarker(line, '<'); marker {
			stack = append(stack, Hunk{})
			sides = append(sides, sideOurs)
			continue
		}
		if _, marker := splitMarker(line, '|'); marker {
			if err := step(stack, sides, sideBase, i); err != nil {
				return nil, err
			}
			continue
		}
		if _, marker := splitMarker(line, '='); marker {
			if err := step(stack, sides, sideTheirs, i); err != nil {
				return nil, err
			}
			continue
		}
		if _, marker := splitMarker(line, '>'); marker {
			if len(stack) == 0 {
				return nil, fmt.Errorf("line %d: end marker outside a conflict", i+1)
			}
			f.Hunks = append(f.Hunks, stack[len(stack)-1])
			stack = stack[:len(stack)-1]
			sides = sides[:len(sides)-1]
			// Text after a conflict opens a new block.
			f.Blocks = append(f.Blocks, nil)
			continue
		}

		if len(stack) == 0 {
			last := len(f.Blocks) - 1
			f.Blocks[last] = append(f.Blocks[last], line)
			continue
		}
		h := &stack[len(stack)-1]
		switch sides[len(sides)-1] {
		case sideOurs:
			h.Ours = append(h.Ours, line)
		case sideBase:
			h.Base = append(h.Base, line)
		case sideTheirs:
			h.Theirs = append(h.Theirs, line)
		}
	}

	if len(stack) != 0 {
		return nil, fmt.Errorf("%d unterminated conflict(s): the file ends inside a conflict", len(stack))
	}
	return f, nil
}

// step moves the innermost hunk to the given side, rejecting a marker that
// arrives outside a conflict or that re-splits one already split. A hunk has
// exactly one base and one separator, so a second of either means the markers
// do not describe a conflict this parser understands.
//
// The separator is legal after ours, which is the default two-way style, and
// after base, which is diff3's extra section; the base marker is legal only
// after ours, since diff3 puts it before the split.
func step(stack []Hunk, sides []parseSide, to parseSide, line int) error {
	what := "separator"
	if to == sideBase {
		what = "base marker"
	}
	if len(sides) == 0 {
		return fmt.Errorf("line %d: %s outside a conflict", line+1, what)
	}
	last := len(sides) - 1
	switch {
	case to == sideTheirs && sides[last] == sideTheirs:
		return fmt.Errorf("line %d: %s after the conflict was already split", line+1, what)
	case to == sideBase && sides[last] != sideOurs:
		return fmt.Errorf("line %d: %s after the conflict was already split", line+1, what)
	}
	sides[last] = to
	if to == sideBase {
		// Recorded on opening the section rather than on its first line: an
		// add/add conflict has an empty base, which is a base that was read
		// and found blank, not a base that was absent.
		stack[last].HasBase = true
	}
	return nil
}

// splitLines splits s into lines with no trailing newline on any of them, and
// reports whether s ended with one.
func splitLines(s string) ([]string, bool) {
	if s == "" {
		return nil, false
	}
	trailing := strings.HasSuffix(s, "\n")
	if trailing {
		s = s[:len(s)-1]
	}
	return strings.Split(s, "\n"), trailing
}

// Render returns the file content with each hunk replaced by the side its
// choice selects. It errors rather than guessing if there are not exactly as
// many choices as hunks, or if a choice is not one of the known resolutions,
// so a caller bug cannot produce a file with a hunk silently dropped.
func (f *ConflictFile) Render(choices []Resolution) (string, error) {
	if len(choices) != len(f.Hunks) {
		return "", fmt.Errorf("%d hunks but %d choices", len(f.Hunks), len(choices))
	}
	out := make([]string, 0, len(f.Blocks[0])+len(f.Hunks))
	for i, h := range f.Hunks {
		out = append(out, f.Blocks[i]...)
		switch choices[i] {
		case ResolveOurs:
			out = append(out, h.Ours...)
		case ResolveTheirs:
			out = append(out, h.Theirs...)
		case ResolveBoth:
			out = append(out, h.Ours...)
			out = append(out, h.Theirs...)
		default:
			return "", fmt.Errorf("hunk %d: unknown resolution %d", i+1, int(choices[i]))
		}
	}
	out = append(out, f.Blocks[len(f.Hunks)]...)
	s := strings.Join(out, "\n")
	if f.trailingNewline && s != "" {
		s += "\n"
	}
	return s, nil
}

// side returns the lines a hunk's resolution keeps, for showing a preview.
func (h Hunk) side(r Resolution) []string {
	switch r {
	case ResolveOurs:
		return h.Ours
	case ResolveTheirs:
		return h.Theirs
	case ResolveBoth:
		return append(append([]string{}, h.Ours...), h.Theirs...)
	}
	return nil
}

// Preview is the text a resolution would write for one hunk, for showing
// before it is committed to disk.
func (h Hunk) Preview(r Resolution) string {
	return strings.Join(h.side(r), "\n")
}

// Labels for the markers MergeConflict asks git to write. They name the sides
// rather than a branch and a revision, because the two are not knowable from
// inside git in every case - during a rebase the "ours" side is the commit
// being replayed, not HEAD - and because the resolver shows both sides of a
// hunk side by side anyway.
const (
	mergeLabelOurs   = "ours"
	mergeLabelBase   = "base"
	mergeLabelTheirs = "theirs"
)

// conflictExitLimit is the highest exit status `git merge-file` uses to report
// conflicts: it returns the number it found, and only goes above this for a
// real error.
const conflictExitLimit = 127

// MergeConflict returns path's conflicts as the hunks a user can pick a side
// for.
//
// It does not read the working tree file. That file's markers are ambiguous
// whenever the file's own content contains marker-shaped lines, and so are the
// markers git merge-file writes in that file's place, so the hunks are merged
// here instead, from the three index stages git itself recorded. Content that
// would read as a marker is escaped out of the way first; see escapeMarkers.
// The consequence is that a user who hand-edited the conflicted file before
// opening the resolver sees those edits replaced on save.
//
// Reading the stages also gives the common ancestor for free, since
// `git merge-file --diff3` records it. A stage can legitimately be missing -
// add/add has no base, add/delete has nothing on the deleted side - and reads
// as empty rather than failing.
//
// path must be unmerged, i.e. FileStatus.Conflicted must be true for it.
func MergeConflict(dir, path string) (*ConflictFile, error) {
	mu.RLock()
	defer mu.RUnlock()

	blobs, err := conflictBlobs(dir, path)
	if err != nil {
		return nil, err
	}
	if len(blobs) == 0 {
		return nil, fmt.Errorf("%s is not conflicted", path)
	}

	// merge-file wants current, base, other; the index numbers them base
	// first, so they are handed over in the order 2, 1, 3.
	var paths [3]string
	for i, stage := range [...]int{2, 1, 3} {
		body := blobs[stage]
		if strings.ContainsRune(body, 0) {
			// git would refuse this as binary anyway, but only after the
			// escape has already rewritten part of it. Splitting "binary"
			// into lines to pick a side of a conflict is meaningless, so
			// say so instead.
			return nil, fmt.Errorf("%s looks binary, so it has no hunks to resolve; "+
				"resolve it by hand or with `git mergetool`", path)
		}
		f, err := writeMsgFile("merge-"+mergeLabelOf(stage), escapeMarkers(body))
		if err != nil {
			return nil, err
		}
		defer os.Remove(f)
		paths[i] = f
	}

	out, conflicts, err := execMergeFile(dir, "merge-file", "-p", "--diff3",
		"-L", mergeLabelOurs, "-L", mergeLabelBase, "-L", mergeLabelTheirs,
		paths[0], paths[1], paths[2])
	if err != nil {
		return nil, err
	}
	merged, err := ParseConflict(out)
	if err != nil {
		return nil, err
	}
	merged.unescape()
	if err := checkConflictCount(path, merged, conflicts); err != nil {
		return nil, err
	}
	return merged, nil
}

// checkConflictCount refuses a parse that disagrees with git about how many
// conflicts there are, which is the one piece of knowledge here that does not
// come from reading text.
//
// Escaping is what makes the parse reliable in the first place, so after it
// this should never fire. It is kept because it costs nothing and because its
// failure mode is the one thing that must not happen here: a hunk parsed
// wrong resolves to a file with lines of the merge quietly missing from it,
// and a line count that disagrees with git is the only sign of that this can
// notice on its own.
func checkConflictCount(path string, merged *ConflictFile, conflicts int) error {
	if len(merged.Hunks) == conflicts {
		return nil
	}
	return fmt.Errorf("%s: git reports %d conflict(s) but the merged text parses as %d; "+
		"refusing to write it, resolve by hand or with `git mergetool`",
		path, conflicts, len(merged.Hunks))
}

// markerEscape is prepended to a line of file content that is shaped like a
// conflict marker, so that a merge of the file produces markers that are
// unambiguously git's own. See escapeMarkers.
//
// It has to be a byte git will still treat as text: NUL makes `git merge-file`
// refuse the file outright as binary, which would leave the escape useless
// exactly where it is most needed.
const markerEscape = "\x01"

// isMarkerLine reports whether line would be read as one of git's markers.
func isMarkerLine(line string) bool {
	for _, c := range []byte{'<', '>', '|', '='} {
		if _, marker := splitMarker(line, c); marker {
			return true
		}
	}
	return false
}

// escapeMarkers prefixes every marker-shaped line of content - and every line
// already starting with markerEscape - with markerEscape.
//
// The markers in a merged file are otherwise ambiguous: given a file whose own
// unchanged text already held a `<<<<<<<` and a `=======`, git merge-file
// reproduces those lines verbatim and writes the real conflict between them, so
// the two run together and nothing in the text says which lines are structure
// and which are content. Counting markers then reports a conflict that is not
// there, one whose half has swallowed the surrounding text - and resolving it
// "theirs" would silently delete that text from the file.
//
// Escaping cannot be reasoned away, only routed around: with content marked,
// every marker left in the merge output is a real one.
//
// The transform is injective and its inverse is well defined. Escaping a line
// that already starts with markerEscape is what guarantees that: afterwards no
// line of the merge can begin with exactly one markerEscape unless it was
// escaped from a line that began with none, so stripping a single leading
// markerEscape restores every line and touches no other.
func escapeMarkers(content string) string {
	if !strings.Contains(content, markerEscape) && !strings.Contains(content, "<<<<<<<") &&
		!strings.Contains(content, ">>>>>>>") && !strings.Contains(content, "=======") &&
		!strings.Contains(content, "|||||||") {
		return content
	}
	lines, trailing := splitLines(content)
	for i, line := range lines {
		if strings.HasPrefix(line, markerEscape) || isMarkerLine(line) {
			lines[i] = markerEscape + line
		}
	}
	out := strings.Join(lines, "\n")
	if trailing {
		out += "\n"
	}
	return out
}

// unescape restores the content of a parsed file, undoing escapeMarkers. Only
// the lines that are content reach a ConflictFile - the markers are consumed
// while parsing - so there is nothing here that must not be unescaped.
func (f *ConflictFile) unescape() {
	for i, block := range f.Blocks {
		f.Blocks[i] = unescapeLines(block)
	}
	for i, h := range f.Hunks {
		f.Hunks[i].Ours = unescapeLines(h.Ours)
		f.Hunks[i].Base = unescapeLines(h.Base)
		f.Hunks[i].Theirs = unescapeLines(h.Theirs)
	}
}

func unescapeLines(lines []string) []string {
	for i, line := range lines {
		lines[i] = strings.TrimPrefix(line, markerEscape)
	}
	return lines
}

// mergeLabelOf names the temp file for one index stage.
func mergeLabelOf(stage int) string {
	switch stage {
	case 1:
		return mergeLabelBase
	case 2:
		return mergeLabelOurs
	case 3:
		return mergeLabelTheirs
	}
	return "unknown"
}

// conflictBlobs returns the content of path's unmerged index stages, keyed by
// stage number. Stage 1 is the common ancestor, 2 the HEAD side and 3 the
// merged-in side.
//
// The caller must hold mu. execGit is called directly for that reason: mu is
// not reentrant, so the read-locking runGit would deadlock against a waiting
// writer rather than simply being redundant.
func conflictBlobs(dir, path string) (map[int]string, error) {
	out, err := execGit(dir, nil, "ls-files", "-u", "-z", "--", path)
	if err != nil {
		return nil, err
	}

	blobs := map[int]string{}
	for _, rec := range strings.Split(out, "\x00") {
		if rec == "" {
			continue
		}
		// Each record is "<mode> <sha> <stage>\t<path>".
		tab := strings.IndexByte(rec, '\t')
		if tab < 0 {
			continue
		}
		field := strings.Fields(rec[:tab])
		if len(field) != 3 {
			continue
		}
		stage, err := strconv.Atoi(field[2])
		if err != nil {
			continue
		}
		content, err := execGit(dir, nil, "cat-file", "-p", field[1])
		if err != nil {
			return nil, err
		}
		blobs[stage] = content
	}
	return blobs, nil
}

// execMergeFile runs `git merge-file`, whose exit status is the number of
// conflicts it found rather than a pass or fail flag, and which writes the
// merged text to stdout precisely when it reports conflicts. execGit discards
// the output of any command that exits non-zero, which would throw away the
// only thing wanted here, so this keeps the output and reads the status, and
// returns the conflict count alongside.
//
// Only a status above conflictExitLimit is a real error.
func execMergeFile(dir string, args ...string) (string, int, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		if err != nil {
			return "", 0, err
		}
		return string(out), 0, nil
	}
	if exitErr.ExitCode() > conflictExitLimit {
		return "", 0, gitError(err, exitErr.Stderr)
	}
	return string(out), exitErr.ExitCode(), nil
}

// ResolveFile writes path with each of its conflicts resolved to the side the
// matching choice selects, then stages the result. Staging is what marks the
// path resolved: `git checkout --ours` leaves it unmerged on its own, and
// `git commit` refuses until every conflicted path has been staged.
//
// choices must have one entry per conflict, in the order MergeConflict
// returned them.
//
// The whole operation holds the write lock: it rewrites a working tree file
// and the index, and a concurrent stage from another goroutine would race it.
// execGit is called directly for that reason - see conflictBlobs.
func ResolveFile(dir, path string, choices []Resolution) error {
	parsed, err := MergeConflict(dir, path)
	if err != nil {
		return err
	}
	if len(parsed.Hunks) == 0 {
		return fmt.Errorf("%s has no conflicts to resolve", path)
	}
	resolved, err := parsed.Render(choices)
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	full := filepath.Join(dir, path)
	// An unmerged path already has a file unless both sides deleted it, and a
	// conflicted script is still an executable one, so keep whatever mode it
	// had rather than falling back to 0644 unconditionally.
	mode := os.FileMode(0o644)
	if info, err := os.Stat(full); err == nil {
		mode = info.Mode().Perm()
	}
	if err := os.WriteFile(full, []byte(resolved), mode); err != nil {
		return err
	}
	_, err = execGit(dir, nil, "add", "--", path)
	return err
}
