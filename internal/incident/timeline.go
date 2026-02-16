package incident

import (
	"github.com/felipemonteiro/mintlog/internal/storage/postgres/queries"
)

func toTimelineEntries(items []queries.IncidentTimeline) []TimelineEntry {
	entries := make([]TimelineEntry, len(items))
	for i, item := range items {
		entries[i] = TimelineEntry{
			ID:        item.ID,
			EventType: item.EventType,
			Content:   item.Content,
			CreatedAt: item.CreatedAt,
		}
	}
	return entries
}
