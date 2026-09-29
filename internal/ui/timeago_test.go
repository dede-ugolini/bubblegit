package ui

import (
	"testing"
	"time"
)

func TestTimeAgo(t *testing.T) {
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		commit time.Time
		want   string
	}{
		{"zero time", time.Time{}, ""},
		{"seconds", now.Add(-30 * time.Second), "1m"},
		{"one minute", now.Add(-time.Minute), "1m"},
		{"three minutes", now.Add(-3 * time.Minute), "3m"},
		{"hours", now.Add(-5*time.Hour - 50*time.Minute), "5h"},
		{"days", now.Add(-2 * 24 * time.Hour), "2d"},
		{"weeks", now.Add(-3 * 7 * 24 * time.Hour), "3w"},
		{"months", now.Add(-4 * 30 * 24 * time.Hour), "4mo"},
		{"just under a year", now.Add(-11 * 30 * 24 * time.Hour), "11mo"},
		{"years", now.Add(-2 * 365 * 24 * time.Hour), "2y"},
		{"future clock skew", now.Add(time.Hour), "1m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := timeAgo(tt.commit, now); got != tt.want {
				t.Fatalf("timeAgo() = %q, want %q", got, tt.want)
			}
		})
	}
}
