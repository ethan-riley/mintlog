package tenant

import "github.com/google/uuid"

type Info struct {
	ID            uuid.UUID
	Name          string
	Plan          string
	RetentionDays int32
	Scopes        []string
	RateLimit     int32
}
