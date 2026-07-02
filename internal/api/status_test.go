package api

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "zero duration",
			duration: 0,
			want:     "0m",
		},
		{
			name:     "minutes only",
			duration: 42 * time.Minute,
			want:     "42m",
		},
		{
			name:     "exactly one hour",
			duration: 1 * time.Hour,
			want:     "1h 0m",
		},
		{
			name:     "hours and minutes",
			duration: 3*time.Hour + 25*time.Minute,
			want:     "3h 25m",
		},
		{
			name:     "exactly one day",
			duration: 24 * time.Hour,
			want:     "1d 0h 0m",
		},
		{
			name:     "days hours minutes",
			duration: 2*24*time.Hour + 5*time.Hour + 30*time.Minute,
			want:     "2d 5h 30m",
		},
		{
			name:     "many days",
			duration: 365 * 24 * time.Hour,
			want:     "365d 0h 0m",
		},
		{
			name:     "59 minutes (no hours)",
			duration: 59 * time.Minute,
			want:     "59m",
		},
		{
			name:     "23 hours 59 minutes (no days)",
			duration: 23*time.Hour + 59*time.Minute,
			want:     "23h 59m",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatDuration(tc.duration)
			if got != tc.want {
				t.Errorf("formatDuration(%v) = %q; want %q", tc.duration, got, tc.want)
			}
		})
	}
}
