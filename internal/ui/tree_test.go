package ui

import (
	"reflect"
	"testing"

	"bubblegit/internal/git"
)

func testFiles(paths ...string) []git.FileStatus {
	files := make([]git.FileStatus, len(paths))
	for i, p := range paths {
		files[i] = git.FileStatus{Index: ' ', Worktree: 'M', Path: p}
	}
	return files
}

func rowNames(rows []fileRow) []string {
	names := make([]string, 0, len(rows))
	for _, r := range rows {
		names = append(names, r.name)
	}
	return names
}

func dirRowNames(rows []fileRow) []string {
	names := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.isDir {
			names = append(names, "[d]"+r.name)
			continue
		}
		names = append(names, r.name)
	}
	return names
}

func TestBuildFileRowsEmpty(t *testing.T) {
	rows := buildFileRows(nil, map[string]bool{})
	if len(rows) != 0 {
		t.Fatalf("expected no rows, got %v", rowNames(rows))
	}
}

func TestBuildFileRowsRootFile(t *testing.T) {
	rows := buildFileRows(testFiles("main.go"), map[string]bool{})
	want := []fileRow{{depth: 0, name: "main.go", fileIdx: 0}}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("got %+v, want %+v", rows, want)
	}
}

func TestBuildFileRowsExpandedByDefault(t *testing.T) {
	rows := buildFileRows(testFiles("a.txt", "dir/b.txt"), map[string]bool{})
	if got, want := dirRowNames(rows), []string{"[d]dir", "b.txt", "a.txt"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if rows[0].isDir != true || rows[0].dir != "dir" || rows[0].depth != 1 {
		t.Fatalf("dir row wrong: %+v", rows[0])
	}
}

func TestBuildFileRowsCollapsed(t *testing.T) {
	rows := buildFileRows(testFiles("dir/b.txt", "a.txt"), map[string]bool{"dir": true})
	if got, want := dirRowNames(rows), []string{"[d]dir", "a.txt"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if rows[1].fileIdx != 1 || rows[1].depth != 0 {
		t.Fatalf("root file row wrong: %+v", rows[1])
	}
}

func TestBuildFileRowsOrderingAndDepth(t *testing.T) {
	files := testFiles("z.txt", "a/b/c.go", "a/b.go", "b/x.go")
	rows := buildFileRows(files, map[string]bool{})
	if got, want := dirRowNames(rows), []string{"[d]a", "[d]b", "c.go", "b.go", "[d]b", "x.go", "z.txt"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	wantDepth := []int{1, 2, 2, 1, 1, 1, 0}
	for i, d := range wantDepth {
		if rows[i].depth != d {
			t.Fatalf("row %d (%s) depth %d, want %d", i, rows[i].name, rows[i].depth, d)
		}
	}
}

func TestBuildFileRowsFileIdxAfterSort(t *testing.T) {
	files := testFiles("b.txt", "dir/z.go", "dir/a.go", "c.txt")
	rows := buildFileRows(files, map[string]bool{})
	got := []string{}
	for _, r := range rows {
		if r.isDir {
			got = append(got, "[d]"+r.name)
			continue
		}
		got = append(got, r.name+files[r.fileIdx].Path)
	}
	want := []string{"[d]dir", "a.godir/a.go", "z.godir/z.go", "b.txtb.txt", "c.txtc.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestBuildFileRowsPartialExpansion(t *testing.T) {
	files := testFiles("a/x.go", "a/b/y.go", "c.go")
	rows := buildFileRows(files, map[string]bool{"a/b": true})
	// "a" stays expanded by default; "a/b" is collapsed.
	if got, want := dirRowNames(rows), []string{"[d]a", "[d]b", "x.go", "c.go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
