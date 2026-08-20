package git

import (
	"reflect"
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
			name:    "empty input",
			input:   "",
			want:    nil,
			wantErr: true,
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
