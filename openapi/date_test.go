package openapi

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDateMarshalJSON(t *testing.T) {
	d := Date{Time: time.Date(2026, time.August, 13, 15, 4, 5, 0, time.UTC)}

	got, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if want := `"2026-08-13"`; string(got) != want {
		t.Errorf("Marshal(%v) = %s, want %s", d, got, want)
	}
}

func TestDateUnmarshalJSON(t *testing.T) {
	var d Date
	if err := json.Unmarshal([]byte(`"2026-08-13"`), &d); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	want := time.Date(2026, time.August, 13, 0, 0, 0, 0, time.UTC)
	if !d.Equal(want) {
		t.Errorf("Unmarshal produced %v, want %v", d.Time, want)
	}
}

func TestDateUnmarshalJSON_Invalid(t *testing.T) {
	tests := []string{
		`"not-a-date"`,
		`"2026-08-13T15:04:05Z"`,
		`123`,
	}

	for _, in := range tests {
		var d Date
		if err := json.Unmarshal([]byte(in), &d); err == nil {
			t.Errorf("Unmarshal(%s) succeeded, want error", in)
		}
	}
}

func TestDateRoundTrip(t *testing.T) {
	original := Date{Time: time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got Date
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !got.Equal(original.Time) {
		t.Errorf("round trip produced %v, want %v", got.Time, original.Time)
	}
}

func TestDateInStruct(t *testing.T) {
	type payload struct {
		D Date `json:"d"`
	}

	data, err := json.Marshal(payload{D: Date{Time: time.Date(2026, time.August, 13, 0, 0, 0, 0, time.UTC)}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if want := `{"d":"2026-08-13"}`; string(data) != want {
		t.Errorf("Marshal = %s, want %s", data, want)
	}

	var got payload
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.D.Format("2006-01-02") != "2026-08-13" {
		t.Errorf("Unmarshal produced %v, want 2026-08-13", got.D.Time)
	}
}

func TestDateTimeMarshalJSON(t *testing.T) {
	dt := DateTime{Time: time.Date(2026, time.August, 13, 15, 4, 5, 0, time.UTC)}

	got, err := json.Marshal(dt)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if want := `"2026-08-13T15:04:05Z"`; string(got) != want {
		t.Errorf("Marshal(%v) = %s, want %s", dt, got, want)
	}
}

func TestDateTimeUnmarshalJSON(t *testing.T) {
	var dt DateTime
	if err := json.Unmarshal([]byte(`"2026-08-13T15:04:05Z"`), &dt); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	want := time.Date(2026, time.August, 13, 15, 4, 5, 0, time.UTC)
	if !dt.Equal(want) {
		t.Errorf("Unmarshal produced %v, want %v", dt.Time, want)
	}
}

func TestDateTimeRoundTrip(t *testing.T) {
	original := DateTime{Time: time.Date(2026, time.August, 13, 15, 4, 5, 123000000, time.UTC)}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got DateTime
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !got.Equal(original.Time) {
		t.Errorf("round trip produced %v, want %v", got.Time, original.Time)
	}
}
