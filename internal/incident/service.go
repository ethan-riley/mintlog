package incident

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/felipemonteiro/mintlog/internal/storage/postgres/queries"
)

type Service struct {
	queries *queries.Queries
}

func NewService(q *queries.Queries) *Service {
	return &Service{queries: q}
}

func (s *Service) Create(ctx context.Context, tenantID uuid.UUID, title, severity string, alertRuleID pgtype.UUID) (queries.Incident, error) {
	if severity == "" {
		severity = "medium"
	}
	return s.queries.CreateIncident(ctx, tenantID, title, StatusTriggered, severity, alertRuleID)
}

func (s *Service) Get(ctx context.Context, id, tenantID uuid.UUID) (queries.Incident, error) {
	return s.queries.GetIncident(ctx, id, tenantID)
}

func (s *Service) List(ctx context.Context, tenantID uuid.UUID, status string) ([]queries.Incident, error) {
	if status != "" {
		return s.queries.ListIncidentsByStatus(ctx, tenantID, status)
	}
	return s.queries.ListIncidents(ctx, tenantID)
}

func (s *Service) UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, newStatus string) (queries.Incident, error) {
	// Validate status transition
	inc, err := s.queries.GetIncident(ctx, id, tenantID)
	if err != nil {
		return queries.Incident{}, err
	}

	if !validTransition(inc.Status, newStatus) {
		return queries.Incident{}, fmt.Errorf("invalid transition from %s to %s", inc.Status, newStatus)
	}

	updated, err := s.queries.UpdateIncidentStatus(ctx, id, tenantID, newStatus)
	if err != nil {
		return queries.Incident{}, err
	}

	// Add timeline entry for status change
	_, _ = s.queries.AddTimelineEntry(ctx, id, "status_change", fmt.Sprintf("Status changed from %s to %s", inc.Status, newStatus))

	return updated, nil
}

func (s *Service) AddTimeline(ctx context.Context, incidentID uuid.UUID, eventType, content string) (queries.IncidentTimeline, error) {
	return s.queries.AddTimelineEntry(ctx, incidentID, eventType, content)
}

func (s *Service) GetTimeline(ctx context.Context, incidentID uuid.UUID) ([]queries.IncidentTimeline, error) {
	return s.queries.GetIncidentTimeline(ctx, incidentID)
}

func validTransition(from, to string) bool {
	switch from {
	case StatusTriggered:
		return to == StatusAcknowledged || to == StatusResolved
	case StatusAcknowledged:
		return to == StatusResolved
	default:
		return false
	}
}
