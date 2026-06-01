package utils

import (
	"fmt"
	"time"
)

// ToUnixMillis converts t to milliseconds since Unix epoch.
func ToUnixMillis(t time.Time) int64 { return t.UnixNano() / int64(time.Millisecond) }

// FromUnixMillis converts milliseconds since Unix epoch to a UTC time.
func FromUnixMillis(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond)).UTC()
}

// Since returns a human-readable string describing the elapsed time since t.
func Since(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}
