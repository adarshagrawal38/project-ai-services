package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	in := []string{"a", "", "b", ""}

	got := Filter(in, func(s string) bool { return s != "" })

	assert.Equal(t, []string{"a", "b"}, got)
	assert.Equal(t, []string{"a", "", "b", ""}, in, "input must not be modified")
}

func TestFilter_EmptyResultIsNotNil(t *testing.T) {
	got := Filter([]int(nil), func(int) bool { return true })

	assert.NotNil(t, got)
	assert.Empty(t, got)
}

func TestFormatTime(t *testing.T) {
	assert.Equal(t, "", FormatTime(time.Time{}))
	assert.Equal(t, "2026-07-01T10:00:00Z", FormatTime(time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)))
}
