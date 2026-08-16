// Hand-written. NOT generated: UPDATE_GOLDEN=1 never touches this file.
// See CLAUDE.md.
package generated

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func serve(t *testing.T, status int, body string) *Client {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)

	return NewClient(srv.URL, nil)
}

// TestListReportsDeclared5xx covers a documented 5xx with a schema — the
// first fixture in the tree to have one.
func TestListReportsDeclared5xx(t *testing.T) {
	for _, status := range []int{http.StatusInternalServerError, http.StatusServiceUnavailable} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			c := serve(t, status, `{"code":7,"message":"upstream down"}`)

			resp, err := c.ListReports(t.Context())
			if resp != nil {
				t.Fatalf("want nil response, got %+v", resp)
			}

			var se ServerError
			if !errors.As(err, &se) {
				t.Fatalf("errors.As(ServerError) failed for %T: %v", err, err)
			}
			if se.Message != "upstream down" || se.Code != 7 {
				t.Errorf("payload = %+v", se)
			}
		})
	}
}

func TestListReportsUndocumentedStatusIsEnvelope(t *testing.T) {
	// 418 is documented nowhere, and ListReports declares no default.
	c := serve(t, http.StatusTeapot, `{"reports":[]}`)

	resp, err := c.ListReports(t.Context())
	if resp != nil {
		t.Fatalf("an undocumented status produced a success value: %+v", resp)
	}

	var ae *APIError
	if !errors.As(err, &ae) || ae.StatusCode != http.StatusTeapot {
		t.Fatalf("err = %v, want *APIError(418)", err)
	}
	var se ServerError
	if errors.As(err, &se) {
		t.Error("an undocumented status must not decode as a documented one")
	}
}

// TestCreateReportDefaultResponse covers the "default" response, which the
// generator previously emitted a type for and then never referenced.
func TestCreateReportDefaultResponse(t *testing.T) {
	c := serve(t, http.StatusTeapot, `{"code":42,"message":"teapot"}`)

	resp, err := c.CreateReport(t.Context(), CreateReportRequest{})
	if resp != nil {
		t.Fatalf("want nil response, got %+v", resp)
	}

	var se ServerError
	if !errors.As(err, &se) {
		t.Fatalf("the default response did not decode; err = %T: %v", err, err)
	}
	if se.Message != "teapot" || se.Code != 42 {
		t.Errorf("payload = %+v", se)
	}
}

func TestCreateReportDefaultDoesNotSwallowSuccess(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusCreated} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			c := serve(t, status, `{"id":"r1","title":"t"}`)

			resp, err := c.CreateReport(t.Context(), CreateReportRequest{})
			if err != nil {
				t.Fatalf("CreateReport: %v", err)
			}
			// Both declared 2xx codes share the Report schema, so both decode.
			if resp.Id != "r1" {
				t.Errorf("resp = %+v", resp)
			}
		})
	}
}

// TestDeleteReportMultiple2xx pins the current behavior for an operation
// declaring both a 200 with a body and a bodiless 204: the first 2xx schema
// in status-code order is the return type, and the 204 decodes to its zero
// value rather than failing on an empty body.
func TestDeleteReportMultiple2xx(t *testing.T) {
	t.Run("200 with body", func(t *testing.T) {
		c := serve(t, http.StatusOK, `{"id":"r1","title":"t"}`)

		resp, err := c.DeleteReport(t.Context(), DeleteReportParams{ReportId: "r1"})
		if err != nil {
			t.Fatalf("DeleteReport: %v", err)
		}
		if resp.Id != "r1" {
			t.Errorf("resp = %+v", resp)
		}
	})

	t.Run("204 no content", func(t *testing.T) {
		c := serve(t, http.StatusNoContent, ``)

		resp, err := c.DeleteReport(t.Context(), DeleteReportParams{ReportId: "r1"})
		if err != nil {
			t.Fatalf("DeleteReport: %v", err)
		}
		if resp == nil || resp.Id != "" {
			t.Errorf("resp = %+v, want the zero value", resp)
		}
	})
}
