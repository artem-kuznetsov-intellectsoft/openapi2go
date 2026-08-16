package openapi

import (
	"encoding/json"
	"time"
)

// dateLayout is the RFC3339 full-date format used by the OpenAPI "date" format.
const dateLayout = "2006-01-02"

// Date holds a calendar date without a time-of-day component, as used by the
// OpenAPI string/date format.
type Date struct {
	time.Time
}

// MarshalJSON renders the date as an RFC3339 full-date string.
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(dateLayout))
}

// MarshalText renders the date as an RFC3339 full-date. Without it, the
// MarshalText promoted from the embedded time.Time would render a Date as a
// full timestamp wherever a text encoding is taken — which is how the client
// runtime formats a date path or query parameter.
func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.Format(dateLayout)), nil
}

// UnmarshalJSON parses an RFC3339 full-date string.
func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return err
	}

	d.Time = t

	return nil
}

// DateTime holds a complete date-time value, as used by the OpenAPI
// string/date-time format. Its JSON encoding is RFC3339 (promoted from the
// embedded time.Time), matching the wire format schema.Format "date-time"
// describes.
type DateTime struct {
	time.Time
}
