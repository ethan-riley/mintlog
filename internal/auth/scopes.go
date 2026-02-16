package auth

const (
	ScopeIngestLogs = "ingest:logs"
	ScopeSearchLogs = "search:logs"
	ScopeAlertRead  = "alerts:read"
	ScopeAlertWrite = "alerts:write"
	ScopeIncidentRead  = "incidents:read"
	ScopeIncidentWrite = "incidents:write"
	ScopeNotifRead  = "notifications:read"
	ScopeNotifWrite = "notifications:write"
	ScopeAdmin      = "admin"
)

var AllScopes = []string{
	ScopeIngestLogs, ScopeSearchLogs,
	ScopeAlertRead, ScopeAlertWrite,
	ScopeIncidentRead, ScopeIncidentWrite,
	ScopeNotifRead, ScopeNotifWrite,
	ScopeAdmin,
}

func HasScope(scopes []string, required string) bool {
	for _, s := range scopes {
		if s == required || s == ScopeAdmin {
			return true
		}
	}
	return false
}
