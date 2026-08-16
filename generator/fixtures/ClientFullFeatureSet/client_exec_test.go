// Hand-written. NOT generated: UPDATE_GOLDEN=1 never touches this file, so a
// regeneration that changes behavior fails here instead of being rubber-
// stamped in a golden diff. See CLAUDE.md.
package generated

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// recorded is what the test server saw, captured before any parsing so the
// assertions are about the bytes on the wire.
type recorded struct {
	method      string
	requestURI  string
	header      http.Header
	pathValue   string
	body        string
	handlerHits int
}

// serve starts a server answering every request with status and body, and
// returns a client pointing at it plus the recorder. Tests that need routing
// itself to be the assertion build their own mux — see
// TestGetItemEscapesPathParameter.
func serve(t *testing.T, status int, body string) (*Client, *recorded) {
	t.Helper()

	got := &recorded{}
	handler := func(w http.ResponseWriter, r *http.Request) {
		got.handlerHits++
		got.method, got.requestURI, got.header = r.Method, r.RequestURI, r.Header.Clone()

		raw, _ := io.ReadAll(r.Body)
		got.body = string(raw)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}

	srv := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(srv.Close)

	return NewClient(srv.URL, nil), got
}

func TestGetItemEscapesPathParameter(t *testing.T) {
	// A trailing slash on the base URL must not produce "//items" either.
	got := &recorded{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{itemId}", func(w http.ResponseWriter, r *http.Request) {
		got.handlerHits++
		got.requestURI, got.pathValue = r.RequestURI, r.PathValue("itemId")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"x","name":"n"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL+"/", nil)
	if _, err := c.GetItem(t.Context(), GetItemParams{XApiKey: "k", ItemId: "a b/c"}); err != nil {
		t.Fatalf("GetItem: %v", err)
	}

	// One segment on the wire, so the {itemId} route matches at all. Before
	// escaping, the raw "/" split this into two segments and the route missed.
	if got.handlerHits != 1 {
		t.Fatalf("handler hit %d times; the URL did not match GET /items/{itemId}", got.handlerHits)
	}
	if want := "/items/a%20b%2Fc"; got.requestURI != want {
		t.Errorf("request-target = %q, want %q", got.requestURI, want)
	}
	if got.pathValue != "a b/c" {
		t.Errorf("server decoded itemId as %q, want %q", got.pathValue, "a b/c")
	}
}

func TestListItemsQueryEncoding(t *testing.T) {
	t.Run("required only", func(t *testing.T) {
		c, got := serve(t, http.StatusOK, `{"total":0}`)
		if _, err := c.ListItems(t.Context(), ListItemsParams{PageSize: 25}); err != nil {
			t.Fatalf("ListItems: %v", err)
		}
		if want := "/items?pageSize=25"; got.requestURI != want {
			t.Errorf("request-target = %q, want %q", got.requestURI, want)
		}
	})

	t.Run("required zero value is still sent", func(t *testing.T) {
		c, got := serve(t, http.StatusOK, `{"total":0}`)
		if _, err := c.ListItems(t.Context(), ListItemsParams{PageSize: 0}); err != nil {
			t.Fatalf("ListItems: %v", err)
		}
		if want := "/items?pageSize=0"; got.requestURI != want {
			t.Errorf("request-target = %q, want %q", got.requestURI, want)
		}
	})

	t.Run("large optional number", func(t *testing.T) {
		c, got := serve(t, http.StatusOK, `{"total":0}`)
		page := float64(1000000)
		if _, err := c.ListItems(t.Context(), ListItemsParams{Page: &page, PageSize: 25}); err != nil {
			t.Fatalf("ListItems: %v", err)
		}
		// fmt.Sprint would have sent page=1e%2B06 here.
		if want := "/items?page=1000000&pageSize=25"; got.requestURI != want {
			t.Errorf("request-target = %q, want %q", got.requestURI, want)
		}
	})
}

func TestGetItemSendsHeaderParams(t *testing.T) {
	t.Run("optional header omitted", func(t *testing.T) {
		c, got := serve(t, http.StatusOK, `{"id":"x"}`)
		if _, err := c.GetItem(t.Context(), GetItemParams{XApiKey: "secret", ItemId: "1"}); err != nil {
			t.Fatalf("GetItem: %v", err)
		}
		if v := got.header.Values("X-Api-Key"); len(v) != 1 || v[0] != "secret" {
			t.Errorf("X-Api-Key = %v, want exactly one \"secret\"", v)
		}
		if _, ok := got.header["X-Request-Id"]; ok {
			t.Error("X-Request-Id was sent for a nil optional parameter")
		}
		if got.header.Get("Content-Type") != "" {
			t.Error("a GET with no body should not set Content-Type")
		}
	})

	t.Run("optional header present", func(t *testing.T) {
		c, got := serve(t, http.StatusOK, `{"id":"x"}`)
		rid := "rid-1"
		_, err := c.GetItem(t.Context(), GetItemParams{XApiKey: "k", XRequestId: &rid, ItemId: "1"})
		if err != nil {
			t.Fatalf("GetItem: %v", err)
		}
		if v := got.header.Values("X-Request-Id"); len(v) != 1 || v[0] != "rid-1" {
			t.Errorf("X-Request-Id = %v", v)
		}
	})
}

func TestCreateItemSendsJSONBody(t *testing.T) {
	c, got := serve(t, http.StatusCreated, `{"id":"1"}`)

	_, err := c.CreateItem(t.Context(), CreateItemRequest{Name: "n", Description: "d"})
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	if ct := got.header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	if want := `{"description":"d","name":"n"}`; got.body != want {
		t.Errorf("body = %q, want %q", got.body, want)
	}
}

func TestGetItemDeclared400ReturnsTypedError(t *testing.T) {
	c, _ := serve(t, http.StatusBadRequest, `{"field":"name","message":"required"}`)

	resp, err := c.GetItem(t.Context(), GetItemParams{XApiKey: "k", ItemId: "1"})
	if resp != nil {
		t.Errorf("want nil response, got %+v", resp)
	}

	var ve ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("errors.As(ValidationError) failed for %T: %v", err, err)
	}
	if ve.Field != "name" || ve.Message != "required" {
		t.Errorf("payload = %+v", ve)
	}

	// Formatting the error used to panic outright.
	if msg := err.Error(); !strings.Contains(msg, "GetItem") || !strings.HasSuffix(msg, "required") {
		t.Errorf("Error() = %q, want it to name the operation and carry the message", msg)
	}
}

// TestUndeclared500IsAnErrorNotASuccess is the sharpest assertion here: the
// body is a payload that unmarshals cleanly into the success type, which is
// exactly how an undeclared status used to be reported as a zero-valued
// success.
func TestUndeclared500IsAnErrorNotASuccess(t *testing.T) {
	c, _ := serve(t, http.StatusInternalServerError, `{"id":"x","name":"n"}`)

	resp, err := c.GetItem(t.Context(), GetItemParams{XApiKey: "k", ItemId: "1"})
	if resp != nil {
		t.Fatalf("a 500 produced a success value: %+v", resp)
	}

	var ae *APIError
	if !errors.As(err, &ae) {
		t.Fatalf("want *APIError, got %T: %v", err, err)
	}
	if ae.StatusCode != http.StatusInternalServerError || ae.Op != "GetItem" {
		t.Errorf("APIError = op %q status %d", ae.Op, ae.StatusCode)
	}

	var ve ValidationError
	if errors.As(err, &ve) {
		t.Error("an undeclared status masqueraded as a declared one")
	}
}

func TestListItemsUndeclaredErrorEnvelopeIsNotSuccess(t *testing.T) {
	// ListItems declares no error responses at all, so it has no switch.
	c, _ := serve(t, http.StatusInternalServerError, `{"error":"internal","code":500}`)

	resp, err := c.ListItems(t.Context(), ListItemsParams{PageSize: 1})
	if resp != nil {
		t.Fatalf("a 500 produced a success value: %+v", resp)
	}
	var ae *APIError
	if !errors.As(err, &ae) || ae.StatusCode != http.StatusInternalServerError {
		t.Fatalf("err = %v, want *APIError(500)", err)
	}
}

func TestCreateItemContentlessErrorCarriesStatus(t *testing.T) {
	c, _ := serve(t, http.StatusUnauthorized, ``)

	_, err := c.CreateItem(t.Context(), CreateItemRequest{})

	var ae *APIError
	if !errors.As(err, &ae) {
		t.Fatalf("want *APIError, got %T: %v", err, err)
	}
	if ae.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d", ae.StatusCode)
	}
	// The old Response401{} marker carried none of this.
	if ae.Op != "CreateItem" || ae.Method != http.MethodPost || ae.URL == "" {
		t.Errorf("envelope = %+v", ae)
	}
}

func TestDeleteItem(t *testing.T) {
	t.Run("204", func(t *testing.T) {
		c, _ := serve(t, http.StatusNoContent, ``)
		if err := c.DeleteItem(t.Context(), DeleteItemParams{ItemId: "1"}); err != nil {
			t.Fatalf("DeleteItem: %v", err)
		}
	})

	t.Run("declared 404", func(t *testing.T) {
		c, _ := serve(t, http.StatusNotFound, `{"resourceId":"item-7"}`)
		err := c.DeleteItem(t.Context(), DeleteItemParams{ItemId: "1"})

		var nf NotFoundError
		if !errors.As(err, &nf) {
			t.Fatalf("errors.As(NotFoundError) failed for %T: %v", err, err)
		}
		if nf.ResourceId != "item-7" {
			t.Errorf("ResourceId = %q", nf.ResourceId)
		}
		// The field-dump fallback: no message-like field on this schema.
		if msg := err.Error(); !strings.Contains(msg, "item-7") {
			t.Errorf("Error() = %q, want it to include the fields", msg)
		}
	})
}

// TestArchiveItemErrorsOnServerError covers the operation that used to end in
// an unconditional "return nil" with no status switch at all, so every status
// including 500 was reported as success.
func TestArchiveItemErrorsOnServerError(t *testing.T) {
	c, _ := serve(t, http.StatusInternalServerError, `boom`)

	err := c.ArchiveItem(t.Context(), ArchiveItemParams{ItemId: "1"})

	var ae *APIError
	if !errors.As(err, &ae) || ae.StatusCode != http.StatusInternalServerError {
		t.Fatalf("ArchiveItem on 500 = %v, want *APIError(500)", err)
	}
}

func TestArchiveItemNoContent(t *testing.T) {
	c, got := serve(t, http.StatusNoContent, ``)

	if err := c.ArchiveItem(t.Context(), ArchiveItemParams{ItemId: "1"}); err != nil {
		t.Fatalf("ArchiveItem: %v", err)
	}
	if want := "/items/1/archive"; got.requestURI != want {
		t.Errorf("request-target = %q, want %q", got.requestURI, want)
	}
	if got.method != http.MethodPatch {
		t.Errorf("method = %q", got.method)
	}
}

func TestReplaceItemConflict(t *testing.T) {
	c, _ := serve(t, http.StatusConflict, `{"message":"version mismatch"}`)

	_, err := c.ReplaceItem(t.Context(), ReplaceItemParams{ItemId: "1"}, ReplaceItemRequest{})

	var ce ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("errors.As(ConflictError) failed for %T: %v", err, err)
	}
	if ce.Message != "version mismatch" {
		t.Errorf("Message = %q", ce.Message)
	}
}

func TestReplaceItemEmptySuccessBody(t *testing.T) {
	c, _ := serve(t, http.StatusOK, ``)

	// An empty 200 body used to fail with "unexpected end of JSON input".
	resp, err := c.ReplaceItem(t.Context(), ReplaceItemParams{ItemId: "1"}, ReplaceItemRequest{})
	if err != nil {
		t.Fatalf("ReplaceItem: %v", err)
	}
	if resp == nil || resp.Id != "" {
		t.Errorf("resp = %+v, want the zero value", resp)
	}
}

func TestNilHTTPClient(t *testing.T) {
	c, _ := serve(t, http.StatusNoContent, ``)

	// serve already passes nil; this asserts it round-trips rather than
	// nil-dereferencing at c.http.Do.
	if err := c.ArchiveItem(t.Context(), ArchiveItemParams{ItemId: "1"}); err != nil {
		t.Fatalf("ArchiveItem: %v", err)
	}
}

func TestRequestOptionApplied(t *testing.T) {
	c, got := serve(t, http.StatusNoContent, ``)

	err := c.ArchiveItem(t.Context(), ArchiveItemParams{ItemId: "1"}, WithBearerToken("tok"))
	if err != nil {
		t.Fatalf("ArchiveItem: %v", err)
	}
	if want := "Bearer tok"; got.header.Get("Authorization") != want {
		t.Errorf("Authorization = %q, want %q", got.header.Get("Authorization"), want)
	}
}

func TestRequestOptionErrorAborts(t *testing.T) {
	c, got := serve(t, http.StatusNoContent, ``)
	sentinel := errors.New("no credential")

	err := c.ArchiveItem(t.Context(), ArchiveItemParams{ItemId: "1"},
		func(*http.Request) error { return sentinel })

	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want the option's error", err)
	}
	if got.handlerHits != 0 {
		t.Error("the request was sent despite the option failing")
	}
}

// TestErrorTypesDoNotPanic is the ten-line test that would have caught the
// panic("TODO: define the output") stubs on day one.
func TestErrorTypesDoNotPanic(t *testing.T) {
	for _, e := range []error{ValidationError{}, ConflictError{}, NotFoundError{}} {
		t.Run(fmt.Sprintf("%T", e), func(t *testing.T) {
			if got := e.Error(); got == "" {
				t.Errorf("Error() = %q, want non-empty", got)
			}
		})
	}
}
