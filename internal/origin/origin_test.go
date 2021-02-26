package origin

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatternCheck(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		origin         string
		url            string
		originPatterns []string
		success        bool
	}{
		{
			name:   "originPatterns",
			origin: "https://two.Example.com",
			url:    "https://example.com/websocket/connection",
			originPatterns: []string{
				"*.example.com",
				"foo.com",
			},
			success: true,
		},
		{
			name:   "originPatternsCyrillicEInOrigin",
			origin: "https://two.еxample.com",
			url:    "https://example.com/websocket/connection",
			originPatterns: []string{
				"*.example.com",
				"foo.com",
			},
			success: false,
		},
		{
			name:   "originPatternsUnauthorized",
			origin: "https://two.Example.com",
			url:    "https://example.com/websocket/connection",
			originPatterns: []string{
				"foo.com",
				"bar.com",
			},
			success: false,
		},
		{
			name:   "fileOrigin",
			origin: "file://",
			url:    "https://example.com/websocket/connection",
			originPatterns: []string{
				"file://*",
			},
			success: true,
		},
		{
			name:   "nullOrigin",
			origin: "null",
			url:    "https://example.com/websocket/connection",
			originPatterns: []string{
				"null",
			},
			success: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest("GET", tc.url, nil)
			r.Header.Set("Origin", tc.origin)

			a, err := NewPatternChecker(tc.originPatterns)
			require.NoError(t, err)
			err = a.Check(r)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCheckSameOrigin(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		origin  string
		url     string
		success bool
	}{
		{
			name:    "no_origin",
			success: true,
			url:     "https://example.com/websocket/connection",
		},
		{
			name:    "invalid_host",
			origin:  "invalid",
			url:     "https://example.com/websocket/connection",
			success: false,
		},
		{
			name:    "unauthorized",
			origin:  "https://example.com",
			url:     "https://example1.com/websocket/connection",
			success: false,
		},
		{
			name:    "authorized",
			origin:  "https://example.com",
			url:     "https://example.com/websocket/connection",
			success: true,
		},
		{
			name:    "authorizedCaseInsensitive",
			origin:  "https://examplE.com",
			url:     "https://example.com/websocket/connection",
			success: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest("GET", tc.url, nil)
			r.Header.Set("Origin", tc.origin)

			err := CheckSameOrigin(r)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}