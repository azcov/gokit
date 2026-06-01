package utils

import "time"

// StartOfDay returns midnight at the beginning of t's day in its location.
func StartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// EndOfDay returns 23:59:59.999999999 at the end of t's day in its location.
func EndOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 23, 59, 59, int(time.Second-1), t.Location())
}

// StartOfMonth returns midnight on the first day of t's month.
func StartOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the last instant of t's month.
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// IsSameDay reports whether a and b fall on the same calendar day (in a's location).
func IsSameDay(a, b time.Time) bool {
	b = b.In(a.Location())
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
