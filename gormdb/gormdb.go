// Package gormdb opens GORM the same way in every Go service, over the service's own connection pool.
//
// The tables come from each service's SQL migrations (platform.Migrate); GORM maps onto them and never creates or
// changes a table. The settings:
//   - times: NowFunc is platform.Now (UTC, to the microsecond), because the columns are timestamp without time zone
//     holding UTC; the models also switch GORM's automatic created_at/updated_at off, so the code sets them
//   - no implicit transaction around each write: repositories open one where it matters
//   - errors are translated, so IsDuplicateKey and IsNotFound work whatever the driver said
//   - slow queries (over 500 ms) and errors go to slog without their values (no user ids or IPs in the log);
//     "record not found" and a taken unique key are answers the caller handles, not errors
package gormdb

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/cinecut-org/cinecut-go-platform/platform"
)

// SlowQuery is how long a query may take before it is logged.
const SlowQuery = 500 * time.Millisecond

// Open puts GORM over pool (the one the migrations ran on), so a service has one pool.
func Open(pool *pgxpool.Pool) (*gorm.DB, error) {
	return gorm.Open(postgres.New(postgres.Config{Conn: stdlib.OpenDBFromPool(pool)}), &gorm.Config{
		NowFunc:                platform.Now,
		SkipDefaultTransaction: true,
		TranslateError:         true,
		Logger: answersNotErrors{logger.NewSlogLogger(slog.Default(), logger.Config{
			SlowThreshold:             SlowQuery,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  logger.Warn,
		})},
	})
}

// answersNotErrors logs a taken unique key as a normal query: the caller turns it into "already exists".
type answersNotErrors struct{ logger.Interface }

func (l answersNotErrors) LogMode(level logger.LogLevel) logger.Interface {
	return answersNotErrors{l.Interface.LogMode(level)}
}

func (l answersNotErrors) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if IsDuplicateKey(err) {
		err = nil
	}
	l.Interface.Trace(ctx, begin, fc, err)
}

// IsNotFound: First, Take or Last found no row.
func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// IsDuplicateKey: a unique key is taken (two writes raced, or the row exists).
func IsDuplicateKey(err error) bool { return errors.Is(err, gorm.ErrDuplicatedKey) }
