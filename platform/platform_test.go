package platform

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetenv(t *testing.T) {
	t.Setenv("CINECUT_TEST", "")
	if Getenv("CINECUT_TEST", "x") != "x" {
		t.Fatal("empty variable should fall back to the default")
	}
	t.Setenv("CINECUT_TEST", "y")
	if Getenv("CINECUT_TEST", "x") != "y" {
		t.Fatal("set variable should win")
	}
}

func TestMustGetenv(t *testing.T) {
	t.Setenv("CINECUT_TEST", "")
	if _, err := MustGetenv("CINECUT_TEST"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	Health("demo")(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"service":"demo"`) {
		t.Fatalf("unexpected response %d %s", rec.Code, rec.Body)
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusBadGateway, "x", "y")
	if rec.Code != 502 || !strings.Contains(rec.Body.String(), `"error":"x"`) {
		t.Fatalf("unexpected response %d %s", rec.Code, rec.Body)
	}
}

func TestIdentityFrom(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set(HeaderUserID, "u1")
	if IdentityFrom(r).UserID != "u1" {
		t.Fatal("user id not read")
	}
}
