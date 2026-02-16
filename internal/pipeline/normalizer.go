package pipeline

import (
	"strings"
	"time"

	"github.com/felipemonteiro/mintlog/pkg/logmodel"
)

var validLevels = map[string]string{
	"trace":   "trace",
	"debug":   "debug",
	"info":    "info",
	"warn":    "warn",
	"warning": "warn",
	"error":   "error",
	"err":     "error",
	"fatal":   "fatal",
	"panic":   "fatal",
}

// NormalizeLevel maps various level strings to a canonical set.
func NormalizeLevel(event *logmodel.LogEvent) {
	lower := strings.ToLower(strings.TrimSpace(event.Level))
	if mapped, ok := validLevels[lower]; ok {
		event.Level = mapped
	} else {
		event.Level = "info"
	}
}

// NormalizeTimestamp ensures the timestamp is set and in UTC.
func NormalizeTimestamp(event *logmodel.LogEvent) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	} else {
		event.Timestamp = event.Timestamp.UTC()
	}
}

// Normalize applies all normalization steps.
func Normalize(event *logmodel.LogEvent) {
	NormalizeTimestamp(event)
	NormalizeLevel(event)
}
