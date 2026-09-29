package ui

import (
	"sort"
	"strings"

	"bubblegit/internal/git"
)

// fileRow is one visible row of the files tree: either a directory or a leaf
// file. fileIdx indexes into m.files and is -1 for directory rows; dir is the
// directory's repo-relative path (used as the expansion key) or empty for
// file rows.
type fileRow struct {
	depth   int
	isDir   bool
	dir     string
	name    string
	fileIdx int
}

// dirNode is a node of the files tree, keyed by its path component relative
// to the parent node.
type dirNode struct {
	dirs   map[string]*dirNode
	files  []int // indices into the surrounding []git.FileStatus
	expKey string
}

// buildFileRows flattens files into the visible rows of a file tree,
// expanding every directory by default and collapsing those whose
// repo-relative path is present in collapsed. Children are ordered
// directories-first then alphabetically.
func buildFileRows(files []git.FileStatus, collapsed map[string]bool) []fileRow {
	root := &dirNode{dirs: make(map[string]*dirNode)}
	for i, f := range files {
		parts := strings.Split(f.Path, "/")
		node := root
		for _, part := range parts[:len(parts)-1] {
			if node.dirs == nil {
				node.dirs = make(map[string]*dirNode)
			}
			child, ok := node.dirs[part]
			if !ok {
				child = &dirNode{}
				node.dirs[part] = child
			}
			node = child
		}
		node.files = append(node.files, i)
	}
	if root.dirs == nil && len(root.files) == 0 {
		return nil
	}

	attachExpKeys(root, "")
	return flattenFileRows(root, files, collapsed)
}

// attachExpKeys walks the tree, recording each directory node's full
// repo-relative path in expKey.
func attachExpKeys(node *dirNode, prefix string) {
	for name, child := range node.dirs {
		child.expKey = name
		if prefix != "" {
			child.expKey = prefix + "/" + name
		}
		attachExpKeys(child, child.expKey)
	}
}

// flattenFileRows DFS-flattens the tree into visible rows, only skipping
// directories present in collapsed. A node's depth in the tree is its number
// of path components (root is depth 0).
func flattenFileRows(node *dirNode, files []git.FileStatus, collapsed map[string]bool) []fileRow {
	var rows []fileRow
	depth := 0
	if node.expKey != "" {
		depth = len(strings.Split(node.expKey, "/"))
	}

	for _, name := range sortedDirNames(node.dirs) {
		child := node.dirs[name]
		rows = append(rows, fileRow{
			depth: depth + 1,
			isDir: true,
			dir:   child.expKey,
			name:  name,
		})
		if !collapsed[child.expKey] {
			rows = append(rows, flattenFileRows(child, files, collapsed)...)
		}
	}

	if len(node.files) > 0 {
		indices := append([]int(nil), node.files...)
		sort.Slice(indices, func(a, b int) bool {
			return filesLess(files[indices[a]].Path, files[indices[b]].Path)
		})
		for _, idx := range indices {
			rows = append(rows, fileRow{
				depth:   depth,
				name:    files[idx].Path[strings.LastIndex(files[idx].Path, "/")+1:],
				fileIdx: idx,
			})
		}
	}

	return rows
}

func sortedDirNames(dirs map[string]*dirNode) []string {
	names := make([]string, 0, len(dirs))
	for name := range dirs {
		names = append(names, name)
	}
	sort.Slice(names, func(a, b int) bool {
		return filesLess(names[a], names[b])
	})
	return names
}

// filesLess orders two repo-relative paths for the tree: files come after
// directories, then alphabetically case-insensitively.
func filesLess(a, b string) bool {
	aDir := strings.Contains(a, "/")
	bDir := strings.Contains(b, "/")
	if aDir != bDir {
		return aDir
	}
	return strings.ToLower(a) < strings.ToLower(b)
}

// rebuildFileTree regenerates the files tree from m.files and m.collapsed,
// clamping the cursor to the new row count.
func (m *Model) rebuildFileTree() {
	m.treeRows = buildFileRows(m.files, m.collapsed)
	m.clampCursors()
}

// selectedFile returns the index into m.files of the file under the cursor,
// and whether it is a file (false when the cursor is on a directory row or
// the tree is empty).
func (m *Model) selectedFile() (int, bool) {
	if len(m.treeRows) == 0 {
		return 0, false
	}
	row := m.treeRows[m.panels[focusStag].idx]
	if row.isDir {
		return 0, false
	}
	return row.fileIdx, true
}
