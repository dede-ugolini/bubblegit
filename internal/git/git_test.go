package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
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

func TestIsRepo(t *testing.T) {
	t.Run("within a repo", func(t *testing.T) {
		dir := initRepo(t)
		if !IsRepo(dir) {
			t.Fatal("IsRepo() = false, want true for a repo root")
		}
	})

	t.Run("subdirectory of a repo", func(t *testing.T) {
		dir := initRepo(t)
		sub := filepath.Join(dir, "sub")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		if !IsRepo(sub) {
			t.Fatal("IsRepo() = false, want true for a repo subdirectory")
		}
	})

	t.Run("not a repo", func(t *testing.T) {
		if IsRepo(t.TempDir()) {
			t.Fatal("IsRepo() = true, want false outside a repository")
		}
	})
}

func TestInit(t *testing.T) {
	headBranch := func(dir string) string {
		cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("failed to get HEAD branch: %v", err)
		}
		return strings.TrimSpace(string(out))
	}

	t.Run("custom branch", func(t *testing.T) {
		dir := t.TempDir()
		if err := Init(dir, "feat"); err != nil {
			t.Fatalf("Init() error: %v", err)
		}
		if !IsRepo(dir) {
			t.Fatal("Init() did not create a repository")
		}
		if got := headBranch(dir); got != "feat" {
			t.Fatalf("HEAD branch = %q, want %q", got, "feat")
		}
	})

	t.Run("git default branch", func(t *testing.T) {
		dir := t.TempDir()
		if err := Init(dir, ""); err != nil {
			t.Fatalf("Init() error: %v", err)
		}
		if !IsRepo(dir) {
			t.Fatal("Init() did not create a repository")
		}
		if got := headBranch(dir); got == "" {
			t.Fatal("Init() left no branch checked out")
		}
	})
}

func TestBranchesCommitTime(t *testing.T) {
	dir := initRepo(t)
	branches, err := Branches(dir)
	if err != nil {
		t.Fatalf("Branches() error: %v", err)
	}
	cur := currentBranch(t, dir)
	for _, b := range branches {
		if b.Name != cur {
			continue
		}
		if b.CommitTime.IsZero() {
			t.Fatalf("CommitTime for %q is zero, want the tip's committer time", cur)
		}
		if d := time.Since(b.CommitTime); d < -time.Minute || d > 2*time.Minute {
			t.Fatalf("CommitTime for %q = %v ago, want within the last 2 minutes", cur, d)
		}
		return
	}
	t.Fatalf("branch %q not found in Branches()", cur)
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
		if err := Merge(dir, "feat", MergeFF); err != nil {
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

		if err := Merge(dir, "feat", MergeCommit); err != nil {
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

		if err := Merge(dir, "feat", MergeSquash); err != nil {
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

func TestDetached(t *testing.T) {
	t.Run("attached to a branch", func(t *testing.T) {
		dir := initRepo(t)

		if hash, ok := Detached(dir); ok {
			t.Errorf("Detached() = %q, true, want not detached", hash)
		}
	})

	t.Run("unborn HEAD", func(t *testing.T) {
		dir := t.TempDir()
		run(t, dir, "git", "init")

		if hash, ok := Detached(dir); ok {
			t.Errorf("Detached() = %q, true, want not detached", hash)
		}
	})

	t.Run("detached at a commit", func(t *testing.T) {
		dir := initRepo(t)
		run(t, dir, "git", "commit", "-q", "--allow-empty", "-m", "second")
		entries, err := Log(dir, "HEAD~1", 1)
		if err != nil {
			t.Fatalf("Log() error: %v", err)
		}
		want := entries[0].ShortHash

		if err := Checkout(dir, "HEAD~1"); err != nil {
			t.Fatalf("Checkout() error: %v", err)
		}

		hash, ok := Detached(dir)
		if !ok {
			t.Fatal("Detached() = not detached, want detached")
		}
		if hash != want {
			t.Errorf("hash = %q, want %q", hash, want)
		}
	})

	t.Run("no branch is current while detached", func(t *testing.T) {
		// The synthetic "(HEAD detached at ...)" row that git emits must not
		// be mistaken for a branch, or the branch panel would offer a
		// checkout target that does not exist.
		dir := initRepo(t)
		run(t, dir, "git", "commit", "-q", "--allow-empty", "-m", "second")
		if err := Checkout(dir, "HEAD~1"); err != nil {
			t.Fatalf("Checkout() error: %v", err)
		}

		branches, err := Branches(dir)
		if err != nil {
			t.Fatalf("Branches() error: %v", err)
		}
		for _, b := range branches {
			if b.Current {
				t.Errorf("branch %q reported current while detached", b.Name)
			}
		}
	})
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

func revParse(t *testing.T, dir, rev string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", rev)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse %s failed: %v", rev, err)
	}
	return strings.TrimSpace(string(out))
}

func revCount(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-list --count failed: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func headMessage(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "log", "-1", "--format=%B")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log -1 --format=%%B failed: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func TestSquashCommits(t *testing.T) {
	t.Run("squash range of commits", func(t *testing.T) {
		dir := initRepo(t)

		if err := writeFile(dir, "hello.go", "line1\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")
		run(t, dir, "git", "commit", "-m", "add line1")

		if err := writeFile(dir, "hello.go", "line1\nline2\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")
		run(t, dir, "git", "commit", "-m", "add line2")
		oldest := revParse(t, dir, "HEAD")

		if err := writeFile(dir, "hello.go", "line1\nline2\nline3\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")
		run(t, dir, "git", "commit", "-m", "add line3")

		if err := SquashCommits(dir, oldest, 2, "squashed: add line2+3"); err != nil {
			t.Fatalf("SquashCommits() error: %v", err)
		}

		if got, want := revCount(t, dir), "3"; got != want {
			t.Errorf("rev-list --count = %q, want %q", got, want)
		}
		if got, want := headMessage(t, dir), "squashed: add line2+3"; got != want {
			t.Errorf("commit message = %q, want %q", got, want)
		}
		content, err := os.ReadFile(filepath.Join(dir, "hello.go"))
		if err != nil {
			t.Fatal(err)
		}
		if got, want := string(content), "line1\nline2\nline3\n"; got != want {
			t.Errorf("hello.go = %q, want %q", got, want)
		}
	})

	t.Run("squash including the root commit", func(t *testing.T) {
		dir := initRepo(t)
		root := revParse(t, dir, "HEAD")

		if err := writeFile(dir, "a.go", "a\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "a.go")
		run(t, dir, "git", "commit", "-m", "add a")

		if err := SquashCommits(dir, root, 2, "squashed root"); err != nil {
			t.Fatalf("SquashCommits() with root error: %v", err)
		}

		if got, want := revCount(t, dir), "1"; got != want {
			t.Errorf("rev-list --count = %q, want %q", got, want)
		}
		if got, want := headMessage(t, dir), "squashed root"; got != want {
			t.Errorf("commit message = %q, want %q", got, want)
		}
	})

	t.Run("fewer than 2 commits fails", func(t *testing.T) {
		dir := initRepo(t)
		if err := SquashCommits(dir, "HEAD", 1, "msg"); err == nil {
			t.Fatal("SquashCommits() with count=1 succeeded, want error")
		}
	})
}

func TestDropLastCommit(t *testing.T) {
	t.Run("drop last commit", func(t *testing.T) {
		dir := initRepo(t)
		if err := writeFile(dir, "hello.go", "line1\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")
		run(t, dir, "git", "commit", "-m", "add line1")

		if err := writeFile(dir, "hello.go", "line1\nline2\n"); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", "hello.go")
		run(t, dir, "git", "commit", "-m", "add line2")

		if got, want := revCount(t, dir), "3"; got != want {
			t.Fatalf("rev-list --count before = %q, want %q", got, want)
		}

		if err := DropLastCommit(dir); err != nil {
			t.Fatalf("DropLastCommit() error: %v", err)
		}

		if got, want := revCount(t, dir), "2"; got != want {
			t.Errorf("rev-list --count after = %q, want %q", got, want)
		}
		if got, want := headMessage(t, dir), "add line1"; got != want {
			t.Errorf("head message = %q, want %q", got, want)
		}
		content, err := os.ReadFile(filepath.Join(dir, "hello.go"))
		if err != nil {
			t.Fatal(err)
		}
		if got, want := string(content), "line1\n"; got != want {
			t.Errorf("hello.go = %q, want %q", got, want)
		}
	})

	t.Run("drop only commit fails", func(t *testing.T) {
		dir := initRepo(t)
		if err := DropLastCommit(dir); err == nil {
			t.Fatal("DropLastCommit() on single commit succeeded, want error")
		}
	})
}

func remoteURL(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git remote get-url origin failed: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func TestTagsByCommit(t *testing.T) {
	t.Run("no tags", func(t *testing.T) {
		dir := initRepo(t)
		tags, err := TagsByCommit(dir)
		if err != nil {
			t.Fatalf("TagsByCommit() error: %v", err)
		}
		if tags != nil {
			t.Fatalf("TagsByCommit() = %v, want nil", tags)
		}
	})

	t.Run("annotated and lightweight tags", func(t *testing.T) {
		dir := initRepo(t)
		hash := revParse(t, dir, "HEAD")

		run(t, dir, "git", "tag", "-a", "v1.0", "-m", "release 1.0")
		run(t, dir, "git", "tag", "light")

		tags, err := TagsByCommit(dir)
		if err != nil {
			t.Fatalf("TagsByCommit() error: %v", err)
		}
		got := tags[hash]
		if len(got) != 2 {
			t.Fatalf("TagsByCommit()[%s] = %v, want 2 tags", hash, got)
		}
		names := map[string]bool{got[0]: true, got[1]: true}
		if !names["v1.0"] || !names["light"] {
			t.Errorf("tags = %v, want v1.0 and light", got)
		}
	})
}

func TestBranchesTracking(t *testing.T) {
	remoteDir := t.TempDir()
	run(t, remoteDir, "git", "init", "--bare", "remote.git")
	remote := filepath.Join(remoteDir, "remote.git")

	dir := initRepo(t)
	main := defaultBranch(t, dir)
	run(t, dir, "git", "remote", "add", "origin", remote)
	run(t, dir, "git", "push", "-u", "origin", main)
	run(t, dir, "git", "checkout", "-b", "feat")
	run(t, dir, "git", "push", "-u", "origin", "feat")
	// solo has no real upstream - its config points at a ref that doesn't
	// exist on the remote, so git reports it as [gone].
	run(t, dir, "git", "branch", "solo")
	run(t, dir, "git", "config", "branch.solo.remote", "origin")
	run(t, dir, "git", "config", "branch.solo.merge", "refs/heads/deleted")

	run(t, dir, "git", "checkout", main)
	if err := writeFile(dir, "hello.go", "package main\n"); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "add", "hello.go")
	run(t, dir, "git", "commit", "-m", "ahead commit")

	got, err := Branches(dir)
	if err != nil {
		t.Fatalf("Branches() error: %v", err)
	}
	byName := map[string]BranchInfo{}
	for _, b := range got {
		byName[b.Name] = b
	}

	cur := byName[main]
	if !cur.Current {
		t.Errorf("current branch %q: Current = false", main)
	}
	if cur.Ahead != 1 {
		t.Errorf("%q ahead = %d, want 1", main, cur.Ahead)
	}
	if cur.Behind != 0 {
		t.Errorf("%q behind = %d, want 0", main, cur.Behind)
	}
	if cur.Upstream != "origin/"+main {
		t.Errorf("%q upstream = %q, want origin/%s", main, cur.Upstream, main)
	}

	if f := byName["feat"]; f.Ahead != 0 || f.Behind != 0 {
		t.Errorf("feat ahead/behind = %d/%d, want 0/0", f.Ahead, f.Behind)
	}

	if s := byName["solo"]; !s.Gone {
		t.Errorf("solo: Gone = false, want true (upstream ref deleted)")
	}
}

func TestAheadHashes(t *testing.T) {
	remoteDir := t.TempDir()
	run(t, remoteDir, "git", "init", "--bare", "remote.git")
	remote := filepath.Join(remoteDir, "remote.git")

	dir := initRepo(t)
	main := defaultBranch(t, dir)
	run(t, dir, "git", "remote", "add", "origin", remote)
	run(t, dir, "git", "push", "-u", "origin", main)

	if err := writeFile(dir, "hello.go", "package main\n"); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "add", "hello.go")
	run(t, dir, "git", "commit", "-m", "ahead commit")
	aheadHash := revParse(t, dir, "HEAD")

	got, err := AheadHashes(dir)
	if err != nil {
		t.Fatalf("AheadHashes() error: %v", err)
	}
	if !got[aheadHash] {
		t.Errorf("AheadHashes() = %v, want it to contain %s", got, aheadHash)
	}
	if len(got) != 1 {
		t.Errorf("AheadHashes() has %d entries, want 1", len(got))
	}

	// An up-to-date branch (pushed, so never ahead) reports nothing.
	run(t, dir, "git", "checkout", "-b", "sync")
	run(t, dir, "git", "push", "-u", "origin", "sync")
	got, err = AheadHashes(dir)
	if err != nil {
		t.Fatalf("AheadHashes() (sync) error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("AheadHashes() (sync) = %v, want empty", got)
	}

	// A branch with no upstream at all also reports nothing.
	run(t, dir, "git", "checkout", "-b", "no-upstream")
	run(t, dir, "git", "commit", "--allow-empty", "-m", "local only")
	got, err = AheadHashes(dir)
	if err != nil {
		t.Fatalf("AheadHashes() (no-upstream) error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("AheadHashes() (no-upstream) = %v, want empty", got)
	}
}

func TestSetRemote(t *testing.T) {
	t.Run("add remote when none exists", func(t *testing.T) {
		dir := initRepo(t)
		if err := SetRemote(dir, "https://example.com/repo.git"); err != nil {
			t.Fatalf("SetRemote() error: %v", err)
		}
		if got, want := remoteURL(t, dir), "https://example.com/repo.git"; got != want {
			t.Errorf("remote url = %q, want %q", got, want)
		}
	})

	t.Run("repoint an existing remote", func(t *testing.T) {
		dir := initRepo(t)
		run(t, dir, "git", "remote", "add", "origin", "https://example.com/old.git")

		if err := SetRemote(dir, "https://example.com/new.git"); err != nil {
			t.Fatalf("SetRemote() error: %v", err)
		}
		if got, want := remoteURL(t, dir), "https://example.com/new.git"; got != want {
			t.Errorf("remote url = %q, want %q", got, want)
		}
	})
}
