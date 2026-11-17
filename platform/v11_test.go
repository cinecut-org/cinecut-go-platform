package platform

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"
	"time"
)

func serve(h http.Handler, headers map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", "/", nil)
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

var ok = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) })

func TestRequireUser(t *testing.T) {
	if c := serve(RequireUser(ok), nil).Code; c != 401 {
		t.Fatalf("anonymous: %d", c)
	}
	if c := serve(RequireUser(ok), map[string]string{HeaderUserID: "not-a-uuid"}).Code; c != 401 {
		t.Fatalf("bad id: %d", c)
	}
	if c := serve(RequireUser(ok), map[string]string{HeaderUserID: "6f1c1a0e-8a5a-4a63-9d55-0c6b1f7d3a10"}).Code; c != 204 {
		t.Fatalf("signed in: %d", c)
	}
}

func TestRequireRole(t *testing.T) {
	h := RequireRole(RoleAdmin)(ok)
	user := "6f1c1a0e-8a5a-4a63-9d55-0c6b1f7d3a10"
	if c := serve(h, nil).Code; c != 401 {
		t.Fatalf("anonymous: %d", c)
	}
	if c := serve(h, map[string]string{HeaderUserID: user, HeaderRoles: RoleContentManager}).Code; c != 403 {
		t.Fatalf("content manager: %d", c)
	}
	if c := serve(h, map[string]string{HeaderUserID: user, HeaderRoles: "CUSTOMER,ADMIN"}).Code; c != 204 {
		t.Fatalf("admin: %d", c)
	}
}

func TestInternalOnly(t *testing.T) {
	if c := serve(InternalOnly("")(ok), map[string]string{HeaderInternalToken: ""}).Code; c != 403 {
		t.Fatalf("unconfigured token must refuse: %d", c)
	}
	if c := serve(InternalOnly("s3cret")(ok), map[string]string{HeaderInternalToken: "wrong"}).Code; c != 403 {
		t.Fatalf("wrong token: %d", c)
	}
	if c := serve(InternalOnly("s3cret")(ok), map[string]string{HeaderInternalToken: "s3cret"}).Code; c != 204 {
		t.Fatalf("right token: %d", c)
	}
}

func TestLocalTimeLooksLikeJavaLocalDateTime(t *testing.T) {
	ts := time.Date(2026, 11, 17, 4, 54, 7, 148123000, time.UTC)
	b, _ := json.Marshal(LocalTime{ts})
	if string(b) != `"2026-11-17T04:54:07.148123"` {
		t.Fatalf("got %s", b)
	}
	b, _ = json.Marshal(LocalTime{ts.Truncate(time.Second)})
	if string(b) != `"2026-11-17T04:54:07"` {
		t.Fatalf("got %s", b)
	}
	var back LocalTime
	if err := json.Unmarshal([]byte(`"2026-11-17T04:54:07.148123"`), &back); err != nil || !back.Equal(ts) {
		t.Fatalf("round trip: %v %v", back, err)
	}
	if p, err := ParseLocalTime("2026-11-17T04:54"); err != nil || !p.Equal(time.Date(2026, 11, 17, 4, 54, 0, 0, time.UTC)) {
		t.Fatalf("Java's toString() form without seconds: %v %v", p, err)
	}
	if p, err := ParseLocalTime("2026-11-17T10:24:07+05:30"); err != nil || !p.Equal(ts.Truncate(time.Second)) {
		t.Fatalf("zoned input should become UTC: %v %v", p, err)
	}
}

func TestNewPage(t *testing.T) {
	p := NewPage[int](nil, 0, 50, 101)
	if p.TotalPages != 3 || p.Items == nil {
		t.Fatalf("%+v", p)
	}
}

// Needs a database: TEST_DATABASE_URL=postgres://cinecut:cinecut@localhost/cinecut_test?sslmode=disable
func TestMigrateAppliesEachFileOnce(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := OpenDB(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	pool.Exec(ctx, `DROP TABLE IF EXISTS schema_migrations, mig_test`) //nolint:errcheck
	files := fstest.MapFS{
		"0001_a.sql": {Data: []byte(`CREATE TABLE mig_test (n int)`)},
		"0002_b.sql": {Data: []byte(`INSERT INTO mig_test VALUES (1)`)},
	}
	for i := 0; i < 2; i++ {
		if err := Migrate(ctx, pool, files); err != nil {
			t.Fatal(err)
		}
	}
	var n int
	pool.QueryRow(ctx, `SELECT count(*) FROM mig_test`).Scan(&n) //nolint:errcheck
	if n != 1 {
		t.Fatalf("0002 ran %d times", n)
	}
	files["0003_bad.sql"] = &fstest.MapFile{Data: []byte(`INSERT INTO nope VALUES (1)`)}
	if err := Migrate(ctx, pool, files); err == nil {
		t.Fatal("a failing migration must be reported")
	}
	pool.Exec(ctx, `DROP TABLE IF EXISTS schema_migrations, mig_test`) //nolint:errcheck
}
