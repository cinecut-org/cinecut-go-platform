package platform

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

// MaxBodyBytes caps a JSON request body.
const MaxBodyBytes = 1 << 20

// WriteJSON writes v as JSON with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// DecodeJSON reads one JSON value into v. Unknown fields are ignored (as Spring does). On failure it has already
// written a 400 and returns false.
func DecodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
	if err := dec.Decode(v); err != nil {
		msg := "The request body is not valid JSON."
		var tooBig *http.MaxBytesError
		switch {
		case errors.As(err, &tooBig):
			msg = "The request body is too large."
		case errors.Is(err, io.EOF):
			msg = "The request body is empty."
		case strings.Contains(err.Error(), "cannot unmarshal"):
			msg = "A field in the request body has the wrong type."
		}
		WriteError(w, http.StatusBadRequest, "INVALID_INPUT", msg)
		return false
	}
	return true
}

// LocalTime is a UTC wall-clock time written like Java's LocalDateTime ("2026-11-17T04:54:07.148123"), which is
// what the monolith's JSON and its "timestamp without time zone" columns use. The monolith's clock is UTC.
type LocalTime struct{ time.Time }

const localLayout = "2006-01-02T15:04:05.999999"

// Now is the current time truncated to microseconds (what Postgres keeps).
func Now() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }

func (t LocalTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.UTC().Format(localLayout))
}

func (t *LocalTime) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsed, err := ParseLocalTime(s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// ParseLocalTime accepts "2026-10-01T10:00:00[.ffffff]" (taken as UTC) or an RFC 3339 time with a zone.
func ParseLocalTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.UTC(), nil
	}
	// Java's LocalDateTime.toString() drops zero seconds ("2026-10-01T10:00"), so accept that too
	for _, layout := range []string{localLayout, "2006-01-02T15:04"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, errors.New("expected a date-time like 2026-10-01T10:00:00")
}

// LocalTimePtr wraps an optional time; nil marshals as null.
func LocalTimePtr(t *time.Time) *LocalTime {
	if t == nil {
		return nil
	}
	return &LocalTime{*t}
}

// Page is the monolith's PageResponse.
type Page[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
}

// NewPage fills the totals like Spring's Page.
func NewPage[T any](items []T, page, size int, total int64) Page[T] {
	if items == nil {
		items = []T{}
	}
	pages := 0
	if size > 0 {
		pages = int((total + int64(size) - 1) / int64(size))
	}
	return Page[T]{Items: items, Page: page, Size: size, TotalItems: total, TotalPages: pages}
}
