package main

import (
	"testing"
	"time"

	"github.com/obzva/snippetbox/internal/assert"
)

func TestPrettifyDate(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
		want  string
	}{
		{
			name:  "UTC",
			input: time.Date(2024, 3, 17, 10, 15, 0, 0, time.UTC),
			want:  "17 Mar 2024 at 10:15",
		},
		{
			name:  "Empty",
			input: time.Time{},
			want:  "",
		},
		{
			name:  "CET",
			input: time.Date(2024, 3, 17, 10, 15, 0, 0, time.FixedZone("CET", 1*60*60)),
			want:  "17 Mar 2024 at 09:15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, prettifyDate(tt.input), tt.want)
		})
	}
}
