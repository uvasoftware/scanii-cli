package terminal

import (
	"strings"
	"testing"
	"time"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		b    uint64
		want string
	}{
		{0, "0 B"},
		{68, "68 B"},
		{999, "999 B"},
		{1000, "1 KB"},
		{1500, "1.5 KB"},
		{1000000, "1 MB"},
		{1200000, "1.2 MB"},
		{1000000000, "1 GB"},
		{1500000000000, "1.5 TB"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatBytes(tt.b)
			if got != tt.want {
				t.Fatalf("FormatBytes(%d) = %q, want %q", tt.b, got, tt.want)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{1500, "1,500"},
		{1000000, "1,000,000"},
		{-1500, "-1,500"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatNumber(tt.n)
			if got != tt.want {
				t.Fatalf("FormatNumber(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{245 * time.Millisecond, "245 ms"},
		{2500 * time.Millisecond, "2.5 s"},
		{1 * time.Second, "1 s"},
		{72 * time.Second, "1.2 min"},
		{5 * time.Minute, "5 min"},
		{3 * time.Hour, "3 h"},
		{25 * time.Hour, "1 d"},
		{500 * time.Microsecond, "500 μs"},
		{48*time.Hour + 12*time.Hour, "2.5 d"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatDuration(tt.d)
			if got != tt.want {
				t.Fatalf("FormatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	// RFC3339Nano input
	result := FormatTime("2026-03-28T10:30:00.123456Z")
	if result == "2026-03-28T10:30:00.123456Z" {
		t.Fatalf("expected formatted time, got raw input: %q", result)
	}
	if !strings.Contains(result, "2026") {
		t.Fatalf("expected year in output, got %q", result)
	}
}

func TestFormatTimeRFC3339(t *testing.T) {
	result := FormatTime("2026-03-28T10:30:00Z")
	if result == "2026-03-28T10:30:00Z" {
		t.Fatalf("expected formatted time, got raw input: %q", result)
	}
}

func TestFormatTimeInvalid(t *testing.T) {
	result := FormatTime("not-a-date")
	if result != "not-a-date" {
		t.Fatalf("expected original string for invalid input, got %q", result)
	}
}
