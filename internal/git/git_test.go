package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []FileStatus
		wantErr bool
	}{
		{
			name:  "modified in worktree",
			input: " M main.go\x00",
			want: []FileStatus{
				{Index: ' ', Worktree: 'M', Path: "main.go"},
			},
		},
		{
			name:  "staged in index",
			input: "M  main.go\x00",
			want: []FileStatus{
				{Index: 'M', Worktree: ' ', Path: "main.go"},
			},
		},
		{
			name:  "modified in both index and worktree",
			input: "MM main.go\x00",
			want: []FileStatus{
				{Index: 'M', Worktree: 'M', Path: "main.go"},
			},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "rename file",
			input: "R  new.go\x00old.go\x00",
			want: []FileStatus{
				{
					Index:    'R',
					Worktree: ' ',
					Path:     "new.go",
					OrigPath: "old.go",
				},
			},
		},
		{
			name:  "multiple files",
			input: " M main.go\x00?? untracked.go\x00",
			want: []FileStatus{
				{Index: ' ', Worktree: 'M', Path: "main.go"},
				{Index: '?', Worktree: '?', Path: "untracked.go"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStatus(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("parseStatus() unexpected error: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseStatus() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseStatus() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	run(t, dir, "git", "commit", "--allow-empty", "-m", "init")
	return dir
}

func defaultBranch(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get default branch: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func TestStatusIntegration(t *testing.T) {
	t.Run("clean repo", func(t *testing.T) {
		dir := initRepo(t)
		got, err := Status(dir)
		if err != nil {
			t.Fatalf("Status() error: %v", err)
		}
		if got != nil {
			t.Fatalf("Status() = %v, want nil", got)
		}
	})

	t.Run("modified file", func(t *testing.T) {
		dir := initRepo(t)
		if err := writeFile(dir, "hello.go", "package main\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")
		run(t, dir, "git", "commit", "-m", "add file")
		if err := writeFile(dir, "hello.go", "package main\n// changed\n"); err != nil {
			t.Fatal(err)
		}

		got, err := Status(dir)
		if err != nil {
			t.Fatalf("Status() error: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("Status() returned %d entries, want 1", len(got))
		}
		if got[0].Worktree != 'M' {
			t.Errorf("Worktree = %c, want M", got[0].Worktree)
		}
	})

	t.Run("untracked file", func(t *testing.T) {
		dir := initRepo(t)
		if err := writeFile(dir, "new.go", "package main\n"); err != nil {
			t.Fatal(err)
		}

		got, err := Status(dir)
		if err != nil {
			t.Fatalf("Status() error: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("Status() returned %d entries, want 1", len(got))
		}
		if got[0].Index != '?' || got[0].Worktree != '?' {
			t.Errorf("got %c%c, want ??", got[0].Index, got[0].Worktree)
		}
	})
}

func TestRestore(t *testing.T) {
	dir := initRepo(t)
	if err := writeFile(dir, "hello.go", "package main\n"); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "add", "hello.go")
	run(t, dir, "git", "commit", "-m", "add hello")

	t.Run("unstaged change", func(t *testing.T) {
		if err := writeFile(dir, "hello.go", "package main\n// unstaged\n"); err != nil {
			t.Fatal(err)
		}
		if err := Restore(dir, "hello.go"); err != nil {
			t.Fatalf("Restore() error: %v", err)
		}
		content, err := os.ReadFile(filepath.Join(dir, "hello.go"))
		if err != nil {
			t.Fatal(err)
		}
		if got := string(content); got != "package main\n" {
			t.Errorf("hello.go = %q, want %q", got, "package main\n")
		}
	})

	t.Run("staged change", func(t *testing.T) {
		if err := writeFile(dir, "hello.go", "package main\n// staged\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")

		status, err := Status(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(status) != 1 || !status[0].Staged() {
			t.Fatalf("expected staged entry, got %#v", status)
		}

		if err := Restore(dir, "hello.go"); err != nil {
			t.Fatalf("Restore() error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(dir, "hello.go"))
		if err != nil {
			t.Fatal(err)
		}
		if got := string(content); got != "package main\n" {
			t.Errorf("hello.go = %q, want %q", got, "package main\n")
		}
		if status, err := Status(dir); err != nil || len(status) != 0 {
			t.Errorf("status after restore = %#v (err=%v), want empty", status, err)
		}
	})

	t.Run("staged and modified", func(t *testing.T) {
		if err := writeFile(dir, "hello.go", "package main\n// staged\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")
		if err := writeFile(dir, "hello.go", "package main\n// staged && worktree\n"); err != nil {
			t.Fatal(err)
		}

		if err := Restore(dir, "hello.go"); err != nil {
			t.Fatalf("Restore() error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(dir, "hello.go"))
		if err != nil {
			t.Fatal(err)
		}
		if got := string(content); got != "package main\n" {
			t.Errorf("hello.go = %q, want %q", got, "package main\n")
		}
		if status, err := Status(dir); err != nil || len(status) != 0 {
			t.Errorf("status after restore = %#v (err=%v), want empty", status, err)
		}
	})
}

func TestMerge(t *testing.T) {
	t.Run("ff", func(t *testing.T) {
		dir := initRepo(t)
		main := defaultBranch(t, dir)
		if err := writeFile(dir, "hello.go", "package main\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")
		run(t, dir, "git", "commit", "-m", "add hello")

		run(t, dir, "git", "checkout", "-b", "feat")
		if err := writeFile(dir, "hello.go", "package main\n// feat\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")
		run(t, dir, "git", "commit", "-m", "feat: change")

		run(t, dir, "git", "checkout", main)
		if err := Merge(dir, "feat", "ff"); err != nil {
			t.Fatalf("Merge(ff) error: %v", err)
		}
		content, err := os.ReadFile(filepath.Join(dir, "hello.go"))
		if err != nil {
			t.Fatal(err)
		}
		if got := string(content); got != "package main\n// feat\n" {
			t.Errorf("hello.go = %q, want feat content", got)
		}
		if got := currentBranch(t, dir); got != main {
			t.Errorf("branch = %q, want %q", got, main)
		}
	})

	t.Run("merge commit", func(t *testing.T) {
		dir := initRepo(t)
		main := defaultBranch(t, dir)
		if err := writeFile(dir, "hello.go", "package main\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")
		run(t, dir, "git", "commit", "-m", "add hello")

		run(t, dir, "git", "checkout", "-b", "feat")
		if err := writeFile(dir, "hello.go", "package main\n// feat\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")
		run(t, dir, "git", "commit", "-m", "feat: change")

		run(t, dir, "git", "checkout", main)
		if err := writeFile(dir, "main.go", "package main\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "main.go")
		run(t, dir, "git", "commit", "-m", "main: change")

		if err := Merge(dir, "feat", "merge"); err != nil {
			t.Fatalf("Merge(merge) error: %v", err)
		}
		cmd := exec.Command("git", "rev-list", "--parents", "-n", "1", "HEAD")
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			t.Fatal(err)
		}
		if parents := len(strings.Fields(string(out))); parents != 3 {
			t.Errorf("merge commit has %d tokens (want 3: HEAD + 2 parents)", parents)
		}
	})

	t.Run("squash", func(t *testing.T) {
		dir := initRepo(t)
		main := defaultBranch(t, dir)
		if err := writeFile(dir, "hello.go", "package main\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")
		run(t, dir, "git", "commit", "-m", "add hello")

		run(t, dir, "git", "checkout", "-b", "feat")
		if err := writeFile(dir, "hello.go", "package main\n// feat\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")
		run(t, dir, "git", "commit", "-m", "feat: change")

		run(t, dir, "git", "checkout", main)
		if err := writeFile(dir, "main.go", "package main\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "main.go")
		run(t, dir, "git", "commit", "-m", "main: change")

		if err := Merge(dir, "feat", "squash"); err != nil {
			t.Fatalf("Merge(squash) error: %v", err)
		}
		cmd := exec.Command("git", "rev-list", "--parents", "-n", "1", "HEAD")
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			t.Fatal(err)
		}
		if parents := len(strings.Fields(string(out))); parents != 2 {
			t.Errorf("squash commit has %d tokens (want 2: HEAD + 1 parent)", parents)
		}
		content, err := os.ReadFile(filepath.Join(dir, "hello.go"))
		if err != nil {
			t.Fatal(err)
		}
		if got := string(content); got != "package main\n// feat\n" {
			t.Errorf("hello.go = %q, want feat content", got)
		}
	})
}

func TestCheckout(t *testing.T) {
	dir := initRepo(t)
	branch := defaultBranch(t, dir)
	run(t, dir, "git", "checkout", "-b", "feat")

	if err := Checkout(dir, branch); err != nil {
		t.Fatalf("Checkout() error: %v", err)
	}

	got := currentBranch(t, dir)
	if got != branch {
		t.Errorf("branch = %q, want %q", got, branch)
	}
}

func TestCreateBranch(t *testing.T) {
	dir := initRepo(t)

	if err := CreateBranch(dir, "new-branch"); err != nil {
		t.Fatalf("CreateBranch() error: %v", err)
	}

	got := currentBranch(t, dir)
	if got != "new-branch" {
		t.Errorf("branch = %q, want %q", got, "new-branch")
	}
}

func TestDeleteBranch(t *testing.T) {
	t.Run("delete other branch", func(t *testing.T) {
		dir := initRepo(t)
		branch := defaultBranch(t, dir)
		run(t, dir, "git", "checkout", "-b", "temp")
		run(t, dir, "git", "checkout", branch)

		if err := DeleteBranch(dir, "temp"); err != nil {
			t.Fatalf("DeleteBranch() error: %v", err)
		}
	})

	t.Run("delete current branch fails", func(t *testing.T) {
		dir := initRepo(t)
		branch := defaultBranch(t, dir)
		err := DeleteBranch(dir, branch)
		if err == nil {
			t.Fatal("DeleteBranch(current) succeeded, want error")
		}
	})
}

func TestCommit(t *testing.T) {
	t.Run("commit staged file", func(t *testing.T) {
		dir := initRepo(t)
		if err := writeFile(dir, "hello.go", "package main\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")

		if err := Commit(dir, "add hello"); err != nil {
			t.Fatalf("Commit() error: %v", err)
		}

		cmd := exec.Command("git", "log", "-1", "--format=%s")
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("failed to read log: %v", err)
		}
		if got := strings.TrimSpace(string(out)); got != "add hello" {
			t.Errorf("commit subject = %q, want %q", got, "add hello")
		}
	})

	t.Run("multi-line message", func(t *testing.T) {
		dir := initRepo(t)
		if err := writeFile(dir, "hello.go", "package main\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")

		if err := Commit(dir, "summary\n\nbody line"); err != nil {
			t.Fatalf("Commit() error: %v", err)
		}

		cmd := exec.Command("git", "log", "-1", "--format=%B")
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("failed to read log: %v", err)
		}
		if got := strings.TrimSpace(string(out)); got != "summary\n\nbody line" {
			t.Errorf("commit body = %q, want %q", got, "summary\n\nbody line")
		}
	})

	t.Run("commit with nothing staged fails", func(t *testing.T) {
		dir := initRepo(t)
		if err := writeFile(dir, "hello.go", "package main\n"); err != nil {
			t.Fatal(err)
		}

		err := Commit(dir, "nothing staged")
		if err == nil {
			t.Fatal("Commit() with nothing staged succeeded, want error")
		}
	})
}

func TestStashBranch(t *testing.T) {
	dir := initRepo(t)
	if err := writeFile(dir, "hello.go", "package main\n"); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "add", "hello.go")
	run(t, dir, "git", "commit", "-m", "add hello")

	if err := writeFile(dir, "hello.go", "package main\n// stashed\n"); err != nil {
		t.Fatal(err)
	}
	if err := StashPush(dir, "wip"); err != nil {
		t.Fatalf("StashPush() error: %v", err)
	}

	if n, err := lenStashEntries(t, dir); err != nil || n != 1 {
		t.Fatalf("stash count = %d, want 1 (err=%v)", n, err)
	}

	if err := StashBranch(dir, "stash-branch", "stash@{0}"); err != nil {
		t.Fatalf("StashBranch() error: %v", err)
	}

	if got := currentBranch(t, dir); got != "stash-branch" {
		t.Errorf("current branch = %q, want %q", got, "stash-branch")
	}

	content, err := os.ReadFile(filepath.Join(dir, "hello.go"))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(content); got != "package main\n// stashed\n" {
		t.Errorf("hello.go = %q, want stashed content", got)
	}

	if n, err := lenStashEntries(t, dir); err != nil || n != 0 {
		t.Fatalf("stash count = %d, want 0 (err=%v)", n, err)
	}
}

func TestPlainDiffBranch(t *testing.T) {
	dir := initRepo(t)
	if err := writeFile(dir, "hello.go", "package main\n"); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "add", "hello.go")
	run(t, dir, "git", "commit", "-m", "add hello")

	if err := writeFile(dir, "hello.go", "package main\n// changed\n"); err != nil {
		t.Fatal(err)
	}

	out, err := DiffBranch(dir)
	if err != nil {
		t.Fatalf("DiffBranch() error: %v", err)
	}
	plain := stripANSI(out)
	if !strings.Contains(plain, "+// changed") {
		t.Errorf("DiffBranch() output missing changed line:\n%s", plain)
	}
	if !strings.Contains(plain, "hello.go") {
		t.Errorf("DiffBranch() output missing file path:\n%s", plain)
	}
}

func TestPlainStashShow(t *testing.T) {
	dir := initRepo(t)
	if err := writeFile(dir, "hello.go", "package main\n"); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "add", "hello.go")
	run(t, dir, "git", "commit", "-m", "add hello")

	if err := writeFile(dir, "hello.go", "package main\n// stashed\n"); err != nil {
		t.Fatal(err)
	}
	if err := StashPush(dir, "wip"); err != nil {
		t.Fatalf("StashPush() error: %v", err)
	}
	if n, err := lenStashEntries(t, dir); err != nil || n != 1 {
		t.Fatalf("stash count = %d, want 1 (err=%v)", n, err)
	}

	out, err := StashShow(dir, "stash@{0}")
	if err != nil {
		t.Fatalf("StashShow() error: %v", err)
	}
	plain := stripANSI(out)
	if !strings.Contains(plain, "+// stashed") {
		t.Errorf("StashShow() output missing stashed line:\n%s", plain)
	}
}

func lenStashEntries(t *testing.T, dir string) (int, error) {
	t.Helper()
	cmd := exec.Command("git", "stash", "list")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return 0, nil
	}
	return len(strings.Split(s, "\n")), nil
}

func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

func currentBranch(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get branch: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func writeFile(dir, name, content string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
}
