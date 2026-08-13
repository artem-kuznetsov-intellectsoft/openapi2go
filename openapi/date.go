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
