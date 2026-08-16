package clientruntime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/url"
	"strings"
	"testing"
	"time"
)

// validationError stands in for a generated error type: a value receiver on
// Error, matching what the generator emits.
type validationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (r validationError) Error() string { return r.Message }

func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()

	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	return NewClient(srv.URL, nil)
}

func TestResolveURL(t *testing.T) {
	tests := []struct {
		name  string
		base  string
		path  string
		query url.Values
		want  string
	}{
		{"plain", "http://h/api", "/items", nil, "http://h/api/items"},
		{"trailing slash on base", "http://h/api/", "/items", nil, "http://h/api/items"},
		{"base path preserved", "http://h/v1", "/items", nil, "http://h/v1/items"},
		{"no path", "http://h/api", "", nil, "http://h/api"},
		{"escaped segment kept", "http://h", "/items/a%2Fb%20c", nil, "http://h/items/a%2Fb%20c"},
		{"query appended", "http://h", "/items", url.Values{"a": {"1"}, "b": {"2"}}, "http://h/items?a=1&b=2"},
		{"empty query omitted", "http://h", "/items", url.Values{}, "http://h/items"},
		{"unterminated brace path", "http://h", "/broken{template", nil, "http://h/broken%7Btemplate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveURL(tt.base, tt.path, tt.query)
			if err != nil {
				t.Fatalf("resolveURL: %v", err)
			}
			if got != tt.want {
				t.Errorf("resolveURL(%q, %q) = %q, want %q", tt.base, tt.path, got, tt.want)
			}
		})
	}
}

func TestPathParam(t *testing.T) {
	tests := []struct{ in, want string }{
		{"plain", "plain"},
		{"a/b", "a%2Fb"},
		{"a b", "a%20b"},
		{"a?b#c", "a%3Fb%23c"},
		{"100%", "100%25"},
		{"ü", "%C3%BC"},
		{"../admin", "..%2Fadmin"},
		{"", ""},
		// PathEscape leaves these alone and url.JoinPath would then resolve
		// them away, retargeting the request at a parent collection.
		{".", "%2E"},
		{"..", "%2E%2E"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := pathParam(tt.in); got != tt.want {
				t.Errorf("pathParam(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestPathParamDotDotCannotEscapeCollection(t *testing.T) {
	got, err := resolveURL("http://h/v1", "/items/"+pathParam(".."), nil)
	if err != nil {
		t.Fatalf("resolveURL: %v", err)
	}
	if want := "http://h/v1/items/%2E%2E"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatValue(t *testing.T) {
	stamp := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		in   any
		want string
	}{
		// fmt.Sprint renders this as "1e+06".
		{"large float64", float64(1000000), "1000000"},
		{"fractional float64", 1.5, "1.5"},
		{"small float64", 0.1, "0.1"},
		{"float32", float32(2.5), "2.5"},
		{"int64", int64(25), "25"},
		{"int32", int32(-3), "-3"},
		{"int", 7, "7"},
		{"string passthrough", "a b", "a b"},
		{"bool", true, "true"},
		{"nil", nil, ""},
		{"bytes", []byte("raw"), "raw"},
		// fmt.Sprint renders this as "2026-08-16 12:00:00 +0000 UTC".
		{"time.Time", stamp, "2026-08-16T12:00:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatValue(tt.in); got != tt.want {
				t.Errorf("formatValue(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDoSetsAcceptAndContentType(t *testing.T) {
	var accept, contentType string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		accept, contentType = r.Header.Get("Accept"), r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := c.do(t.Context(), request{op: "Op", method: http.MethodPost, path: "/x", body: map[string]int{"a": 1}}, nil); err != nil {
		t.Fatalf("do: %v", err)
	}
	if accept != "application/json" {
		t.Errorf("Accept = %q", accept)
	}
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q", contentType)
	}
}

func TestDoOmitsContentTypeWithoutBody(t *testing.T) {
	var contentType string
	var contentLength int64
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		contentType, contentLength = r.Header.Get("Content-Type"), r.ContentLength
		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := c.do(t.Context(), request{op: "Op", method: http.MethodGet, path: "/x"}, nil); err != nil {
		t.Fatalf("do: %v", err)
	}
	if contentType != "" {
		t.Errorf("Content-Type = %q, want empty", contentType)
	}
	if contentLength != 0 {
		t.Errorf("ContentLength = %d, want 0", contentLength)
	}
}

func TestDoSendsCookies(t *testing.T) {
	var got string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if ck, err := r.Cookie("session"); err == nil {
			got = ck.Value
		}
		w.WriteHeader(http.StatusNoContent)
	})

	req := request{
		op: "Op", method: http.MethodGet, path: "/x",
		cookies: []*http.Cookie{{Name: "session", Value: "abc"}},
	}
	if _, err := c.do(t.Context(), req, nil); err != nil {
		t.Fatalf("do: %v", err)
	}
	if got != "abc" {
		t.Errorf("cookie = %q, want %q", got, "abc")
	}
}

// TestDoReusesConnection is the assertion that actually proves the drain fix:
// an operation whose body is never decoded must still return its connection
// to the idle pool.
func TestDoReusesConnection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"pad":%q}`, strings.Repeat("x", 4096))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, &http.Client{Transport: &http.Transport{}})

	if _, err := c.do(t.Context(), request{op: "Op", method: http.MethodGet, path: "/x"}, nil); err != nil {
		t.Fatalf("first call: %v", err)
	}

	var reused bool
	ctx := httptrace.WithClientTrace(t.Context(), &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) { reused = info.Reused },
	})
	if _, err := c.do(ctx, request{op: "Op", method: http.MethodGet, path: "/x"}, nil); err != nil {
		t.Fatalf("second call: %v", err)
	}

	if !reused {
		t.Error("connection was not reused; the response body is not being drained")
	}
}

func TestReadBodyRejectsOversizeResponse(t *testing.T) {
	body := strings.NewReader(strings.Repeat("x", maxResponseBytes+1))
	if _, err := readBody(io.NopCloser(body)); !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("err = %v, want ErrResponseTooLarge", err)
	}
}

func TestNewClientNilHTTPClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	// The obvious call; it used to nil-deref at c.http.Do.
	c := NewClient(srv.URL, nil)
	if _, err := c.do(t.Context(), request{op: "Op", method: http.MethodGet, path: "/x"}, nil); err != nil {
		t.Fatalf("do: %v", err)
	}
}

func TestRequestOptionOrder(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Token")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, nil, WithHeader("X-Token", "client"))
	_, err := c.do(t.Context(), request{op: "Op", method: http.MethodGet, path: "/x"},
		[]RequestOption{WithHeader("X-Token", "percall")})
	if err != nil {
		t.Fatalf("do: %v", err)
	}

	if got != "percall" {
		t.Errorf("X-Token = %q, want the per-call option to win", got)
	}
}

func TestRequestOptionErrorAbortsBeforeSending(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	sentinel := errors.New("token refresh failed")
	c := NewClient(srv.URL, nil, func(*http.Request) error { return sentinel })

	_, err := c.do(t.Context(), request{op: "Op", method: http.MethodGet, path: "/x"}, nil)
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want the option's error", err)
	}
	if hits != 0 {
		t.Errorf("handler was hit %d times; the request should never have been sent", hits)
	}
}

func TestClientLevelOptionsNotMutated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, nil, WithHeader("A", "1"))
	req := request{op: "Op", method: http.MethodGet, path: "/x"}

	for range 2 {
		if _, err := c.do(t.Context(), req, []RequestOption{WithHeader("B", "2")}); err != nil {
			t.Fatalf("do: %v", err)
		}
	}

	if len(c.opts) != 1 {
		t.Errorf("client options grew to %d; per-call options leaked into c.opts", len(c.opts))
	}
}

func TestDecodeJSONEmptyBody(t *testing.T) {
	for _, body := range []string{"", "   \n", "\t"} {
		t.Run(fmt.Sprintf("%q", body), func(t *testing.T) {
			got, err := decodeJSON[validationError](&HTTPResponse{StatusCode: 204, Body: []byte(body)}, "Op")
			if err != nil {
				t.Fatalf("decodeJSON: %v", err)
			}
			if *got != (validationError{}) {
				t.Errorf("got %+v, want zero value", *got)
			}
		})
	}
}

func TestDecodeJSONFailureCarriesContext(t *testing.T) {
	r := &HTTPResponse{StatusCode: 200, Status: "200 OK", Body: []byte(`<html>nope</html>`)}

	_, err := decodeJSON[validationError](r, "GetItem")
	if !errors.Is(err, ErrDecode) {
		t.Fatalf("err = %v, want ErrDecode", err)
	}
	if !strings.Contains(err.Error(), "GetItem") || !strings.Contains(err.Error(), "200") {
		t.Errorf("err = %q, want it to name the operation and status", err)
	}
}

func TestExpectSuccess(t *testing.T) {
	tests := []struct {
		status  int
		wantErr bool
	}{
		{200, false},
		{201, false},
		{204, false},
		{299, false},
		{301, true},
		{400, true},
		{404, true},
		{500, true},
		{503, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.status), func(t *testing.T) {
			r := &HTTPResponse{StatusCode: tt.status, Status: http.StatusText(tt.status)}
			err := r.expectSuccess("Op")
			if (err != nil) != tt.wantErr {
				t.Errorf("expectSuccess(%d) = %v, wantErr %v", tt.status, err, tt.wantErr)
			}
		})
	}
}

func TestExpectSuccessDefault(t *testing.T) {
	t.Run("2xx passes through", func(t *testing.T) {
		r := &HTTPResponse{StatusCode: http.StatusOK, Status: "200 OK"}
		if err := expectSuccessDefault[validationError](r, "Op"); err != nil {
			t.Errorf("expectSuccessDefault = %v, want nil", err)
		}
	})

	t.Run("undocumented status decodes the default schema", func(t *testing.T) {
		r := &HTTPResponse{
			StatusCode: http.StatusTeapot, Status: "418 I'm a teapot",
			Body: []byte(`{"field":"x","message":"teapot"}`),
		}

		err := expectSuccessDefault[validationError](r, "Op")

		var ve validationError
		if !errors.As(err, &ve) {
			t.Fatalf("errors.As failed for %T: %v", err, err)
		}
		if ve.Message != "teapot" {
			t.Errorf("payload = %+v", ve)
		}
	})
}

func TestDecodeErrorRecoversPayload(t *testing.T) {
	r := &HTTPResponse{
		StatusCode: 400,
		Status:     "400 Bad Request",
		method:     http.MethodGet,
		requestURL: "http://h/items/1",
		Body:       []byte(`{"field":"name","message":"required"}`),
	}

	err := decodeError[validationError](r, "GetItem")

	// The generated type is recovered by value, matching its value receiver.
	var ve validationError
	if !errors.As(err, &ve) {
		t.Fatalf("errors.As(%T) failed", err)
	}
	if ve.Field != "name" || ve.Message != "required" {
		t.Errorf("payload = %+v", ve)
	}

	// The envelope stays reachable for status, URL, and raw body.
	var ae *APIError
	if !errors.As(err, &ae) {
		t.Fatal("errors.As(*APIError) failed")
	}
	if ae.StatusCode != http.StatusBadRequest || ae.Op != "GetItem" || ae.URL != "http://h/items/1" {
		t.Errorf("envelope = %+v", ae)
	}

	want := "GetItem: GET http://h/items/1: 400 Bad Request: required"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestDecodeErrorUndecodableBodyKeepsStatus(t *testing.T) {
	r := &HTTPResponse{
		StatusCode: 400, Status: "400 Bad Request",
		method: http.MethodGet, requestURL: "http://h/x",
		Body: []byte(`<html>gateway</html>`),
	}

	err := decodeError[validationError](r, "GetItem")

	// The HTTP status is the real failure; it must not be replaced by a JSON
	// syntax error.
	var ae *APIError
	if !errors.As(err, &ae) || ae.StatusCode != http.StatusBadRequest {
		t.Fatalf("err = %v, want an *APIError carrying 400", err)
	}
	var ve validationError
	if errors.As(err, &ve) {
		t.Error("a body that did not decode should leave no payload")
	}
	if !strings.Contains(err.Error(), "gateway") {
		t.Errorf("Error() = %q, want it to excerpt the body", err)
	}
}

func TestAPIErrorSummarizesLongBody(t *testing.T) {
	r := &HTTPResponse{
		StatusCode: 500, Status: "500 Internal Server Error",
		Body: []byte(strings.Repeat("x", bodySummaryLimit*2)),
	}

	msg := r.apiError("Op").Error()
	if !strings.HasSuffix(msg, "...") {
		t.Errorf("Error() = %q, want a truncated body excerpt", msg)
	}
	if len(msg) > bodySummaryLimit+128 {
		t.Errorf("Error() is %d bytes; the body excerpt is not bounded", len(msg))
	}
}

func TestDoReturnsResponseForServerError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"error":"boom"}`)
	})

	// A 5xx is a response, not a transport failure: do returns it and lets
	// the generated method classify it.
	resp, err := c.do(t.Context(), request{op: "Op", method: http.MethodGet, path: "/x"}, nil)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d", resp.StatusCode)
	}
	if string(resp.Body) != `{"error":"boom"}` {
		t.Errorf("Body = %q", resp.Body)
	}
}

func TestDoTransportErrorIsNotAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	base := srv.URL
	srv.Close()

	c := NewClient(base, nil)
	_, err := c.do(context.Background(), request{op: "Op", method: http.MethodGet, path: "/x"}, nil)
	if err == nil {
		t.Fatal("want an error from a closed server")
	}

	var ae *APIError
	if errors.As(err, &ae) {
		t.Error("a transport failure should not be reported as an *APIError")
	}
	if !strings.Contains(err.Error(), "Op") {
		t.Errorf("err = %q, want it to name the operation", err)
	}
}
