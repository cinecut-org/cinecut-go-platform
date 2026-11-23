package gormdb

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/cinecut-org/cinecut-go-platform/platform"
)

type row struct {
	ID        uuid.UUID `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime:false"`
}

func (row) TableName() string { return "gormdb_check" }

func TestOpenTimesAndErrors(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := platform.OpenDB(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, `DROP TABLE IF EXISTS gormdb_check; CREATE TABLE gormdb_check (id uuid PRIMARY KEY,
		name text NOT NULL UNIQUE, created_at timestamp(6) without time zone NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(ctx, `DROP TABLE gormdb_check`) //nolint:errcheck
	db, err := Open(pool)
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 11, 23, 10, 0, 0, 123456000, time.UTC)
	if err := db.Create(&row{ID: uuid.New(), Name: "a", CreatedAt: at}).Error; err != nil {
		t.Fatal(err)
	}
	var got row
	if err := db.Take(&got, "name = ?", "a").Error; err != nil {
		t.Fatal(err)
	}
	if !got.CreatedAt.Equal(at) || got.CreatedAt.Location() != time.UTC {
		t.Errorf("a UTC time must come back as stored, in UTC: %v", got.CreatedAt)
	}
	if err := db.Create(&row{ID: uuid.New(), Name: "a", CreatedAt: at}).Error; !IsDuplicateKey(err) {
		t.Errorf("a taken unique key: %v", err)
	}
	if err := db.Take(&got, "name = ?", "nobody").Error; !IsNotFound(err) {
		t.Errorf("no row: %v", err)
	}
	if now := db.NowFunc(); now.Location() != time.UTC || now.Nanosecond()%1000 != 0 {
		t.Errorf("GORM's clock must be platform.Now: %v", now)
	}
}
