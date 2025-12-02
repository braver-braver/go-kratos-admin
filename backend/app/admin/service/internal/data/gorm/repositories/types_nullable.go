package repositories

import (
	"database/sql"
	"time"
)

// NullableTime wraps sql.NullTime for DTO usage, keeping JSON friendly zero value handling.
type NullableTime struct {
	sql.NullTime
}

func NewNullableTime(t time.Time) NullableTime {
	return NullableTime{sql.NullTime{Time: t, Valid: true}}
}

func NewNullableTimePtr(t *time.Time) NullableTime {
	if t == nil || t.IsZero() {
		return NullableTime{}
	}
	return NullableTime{sql.NullTime{Time: *t, Valid: true}}
}

// ValueOrZero ValueOr zero time if invalid.
func (n NullableTime) ValueOrZero() time.Time {
	if n.Valid {
		return n.Time
	}
	return time.Time{}
}
