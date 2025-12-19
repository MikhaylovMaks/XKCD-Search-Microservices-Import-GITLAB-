package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   UpdateStatus
		expected string
	}{
		{
			name:     "StatusUpdateUnknown",
			status:   StatusUpdateUnknown,
			expected: "unknown",
		},
		{
			name:     "StatusUpdateIdle",
			status:   StatusUpdateIdle,
			expected: "idle",
		},
		{
			name:     "StatusUpdateRunning",
			status:   StatusUpdateRunning,
			expected: "running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestUpdateStats(t *testing.T) {
	stats := UpdateStats{
		WordsTotal:    100,
		WordsUnique:   50,
		ComicsFetched: 25,
		ComicsTotal:   30,
	}

	assert.Equal(t, 100, stats.WordsTotal)
	assert.Equal(t, 50, stats.WordsUnique)
	assert.Equal(t, 25, stats.ComicsFetched)
	assert.Equal(t, 30, stats.ComicsTotal)
}

func TestComics(t *testing.T) {
	comic := Comics{
		ID:    123,
		URL:   "https://example.com/comic/123",
		Score: 95,
	}

	assert.Equal(t, 123, comic.ID)
	assert.Equal(t, "https://example.com/comic/123", comic.URL)
	assert.Equal(t, 95, comic.Score)
}

func TestComics_ZeroValues(t *testing.T) {
	comic := Comics{}

	assert.Equal(t, 0, comic.ID)
	assert.Equal(t, "", comic.URL)
	assert.Equal(t, 0, comic.Score)
}

func TestComics_NegativeScore(t *testing.T) {
	comic := Comics{
		ID:    456,
		URL:   "https://example.com/comic/456",
		Score: -10,
	}

	assert.Equal(t, 456, comic.ID)
	assert.Equal(t, "https://example.com/comic/456", comic.URL)
	assert.Equal(t, -10, comic.Score)
}

func TestComics_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		comic Comics
	}{
		{
			name: "empty URL",
			comic: Comics{
				ID:    1,
				URL:   "",
				Score: 100,
			},
		},
		{
			name: "very long URL",
			comic: Comics{
				ID:    2,
				URL:   "https://example.com/very/long/path/to/comic/with/many/segments/and/parameters?query=value&another=param",
				Score: 50,
			},
		},
		{
			name: "maximum ID",
			comic: Comics{
				ID:    2147483647, // max int32
				URL:   "https://example.com/max",
				Score: 999999,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.comic.ID, tt.comic.ID)
			assert.Equal(t, tt.comic.URL, tt.comic.URL)
			assert.Equal(t, tt.comic.Score, tt.comic.Score)
		})
	}
}

func TestUpdateStats_Calculations(t *testing.T) {
	stats := UpdateStats{
		WordsTotal:    1000,
		WordsUnique:   750,
		ComicsFetched: 500,
		ComicsTotal:   1000,
	}
	assert.Equal(t, 1000, stats.WordsTotal)
	assert.Equal(t, 750, stats.WordsUnique)
	assert.Equal(t, 500, stats.ComicsFetched)
	assert.Equal(t, 1000, stats.ComicsTotal)
	assert.True(t, stats.WordsTotal >= stats.WordsUnique)
	assert.True(t, stats.ComicsTotal >= stats.ComicsFetched)
}

func TestUpdateStats_ZeroValues(t *testing.T) {
	stats := UpdateStats{}

	assert.Equal(t, 0, stats.WordsTotal)
	assert.Equal(t, 0, stats.WordsUnique)
	assert.Equal(t, 0, stats.ComicsFetched)
	assert.Equal(t, 0, stats.ComicsTotal)
}

func TestUpdateStats_NegativeValues(t *testing.T) {
	stats := UpdateStats{
		WordsTotal:    -100,
		WordsUnique:   -50,
		ComicsFetched: -25,
		ComicsTotal:   -10,
	}

	assert.Equal(t, -100, stats.WordsTotal)
	assert.Equal(t, -50, stats.WordsUnique)
	assert.Equal(t, -25, stats.ComicsFetched)
	assert.Equal(t, -10, stats.ComicsTotal)
}

func TestUpdateStatus_StringConversion(t *testing.T) {
	tests := []struct {
		status   UpdateStatus
		expected string
	}{
		{StatusUpdateUnknown, "unknown"},
		{StatusUpdateIdle, "idle"},
		{StatusUpdateRunning, "running"},
		{"custom", "custom"}, // Test custom status
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestUpdateStatus_Comparison(t *testing.T) {
	assert.Equal(t, StatusUpdateIdle, StatusUpdateIdle)
	assert.Equal(t, StatusUpdateRunning, StatusUpdateRunning)
	assert.Equal(t, StatusUpdateUnknown, StatusUpdateUnknown)

	assert.NotEqual(t, StatusUpdateIdle, StatusUpdateRunning)
	assert.NotEqual(t, StatusUpdateRunning, StatusUpdateUnknown)
	assert.NotEqual(t, StatusUpdateUnknown, StatusUpdateIdle)
}

func TestComicsSlice_Sorting(t *testing.T) {
	comics := []Comics{
		{ID: 1, URL: "url1", Score: 50},
		{ID: 2, URL: "url2", Score: 100},
		{ID: 3, URL: "url3", Score: 25},
	}
	assert.Len(t, comics, 3)
	assert.Equal(t, 50, comics[0].Score)
	assert.Equal(t, 100, comics[1].Score)
	assert.Equal(t, 25, comics[2].Score)
}
func TestUpdateStats_Percentages(t *testing.T) {
	tests := []struct {
		name  string
		stats UpdateStats
	}{
		{
			name: "normal stats",
			stats: UpdateStats{
				WordsTotal:    1000,
				WordsUnique:   500,
				ComicsFetched: 200,
				ComicsTotal:   1000,
			},
		},
		{
			name: "zero fetched",
			stats: UpdateStats{
				WordsTotal:    1000,
				WordsUnique:   500,
				ComicsFetched: 0,
				ComicsTotal:   1000,
			},
		},
		{
			name: "all fetched",
			stats: UpdateStats{
				WordsTotal:    1000,
				WordsUnique:   500,
				ComicsFetched: 1000,
				ComicsTotal:   1000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.stats.ComicsFetched <= tt.stats.ComicsTotal)
			assert.True(t, tt.stats.WordsUnique <= tt.stats.WordsTotal)
		})
	}
}
