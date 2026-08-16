// Hand-written. NOT generated: UPDATE_GOLDEN=1 never touches this file.
// See CLAUDE.md.
package generated

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type seen struct {
	requestURI string
	cookie     string
}

func serve(t *testing.T) (*Client, *seen) {
	t.Helper()

	got := &seen{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.requestURI = r.RequestURI
		if ck, err := r.Cookie("session"); err == nil {
			got.cookie = ck.Value
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total":0}`))
	}))
	t.Cleanup(srv.Close)

	return NewClient(srv.URL, nil), got
}

// TestArrayQueryParamsRepeatKeys covers the encoding that used to render a
// slice as a single "[a b c]" value.
func TestArrayQueryParamsRepeatKeys(t *testing.T) {
	c, got := serve(t)

	params := SearchParams{
		GroupId: "g1",
		Session: "sess",
		Tags:    []string{"alpha", "beta"},
		Ids:     []int64{1, 2},
	}
	if _, err := c.Search(t.Context(), params); err != nil {
		t.Fatalf("Search: %v", err)
	}

	want := "/search/g1?ids=1&ids=2&tags=alpha&tags=beta"
	if got.requestURI != want {
		t.Errorf("request-target = %q, want %q", got.requestURI, want)
	}
}

// TestCookieParamIsSent covers a parameter location the generator resolved
// into the Params struct and then dropped on the floor.
func TestCookieParamIsSent(t *testing.T) {
	c, got := serve(t)

	if _, err := c.Search(t.Context(), SearchParams{GroupId: "g1", Session: "sess-1"}); err != nil {
		t.Fatalf("Search: %v", err)
	}

	if got.cookie != "sess-1" {
		t.Errorf("session cookie = %q, want %q", got.cookie, "sess-1")
	}
}

func TestScalarQueryFormatting(t *testing.T) {
	c, got := serve(t)

	exact := true
	score := float64(1000000)
	since := DateTime{Time: time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)}
	day := Date{Time: time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)}

	params := SearchParams{
		GroupId: "g1", Session: "s",
		Exact: &exact, Score: &score, Since: &since, Day: &day,
	}
	if _, err := c.Search(t.Context(), params); err != nil {
		t.Fatalf("Search: %v", err)
	}

	// score would be "1e%2B06" under fmt.Sprint; since would be Go's default
	// time layout; day would be a full timestamp without Date.MarshalText.
	want := "/search/g1?day=2026-08-16&exact=true&score=1000000&since=2026-08-16T12%3A00%3A00Z"
	if got.requestURI != want {
		t.Errorf("request-target = %q,\n                want %q", got.requestURI, want)
	}
}

func TestEmptySliceOmitsQueryKey(t *testing.T) {
	c, got := serve(t)

	params := SearchParams{GroupId: "g1", Session: "s", Tags: []string{}}
	if _, err := c.Search(t.Context(), params); err != nil {
		t.Fatalf("Search: %v", err)
	}

	if want := "/search/g1"; got.requestURI != want {
		t.Errorf("request-target = %q, want %q — an empty slice must send no key", got.requestURI, want)
	}
}

func TestPathParamEscaped(t *testing.T) {
	c, got := serve(t)

	if _, err := c.Search(t.Context(), SearchParams{GroupId: "a/b", Session: "s"}); err != nil {
		t.Fatalf("Search: %v", err)
	}

	if want := "/search/a%2Fb"; got.requestURI != want {
		t.Errorf("request-target = %q, want %q", got.requestURI, want)
	}
}
