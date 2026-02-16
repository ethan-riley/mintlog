package queries

import (
	"context"

	"github.com/google/uuid"
)

const createNotificationChannel = `
INSERT INTO notification_channels (tenant_id, name, channel_type, config)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, name, channel_type, config, is_active, created_at, updated_at
`

func (q *Queries) CreateNotificationChannel(ctx context.Context, tenantID uuid.UUID, name, channelType string, config []byte) (NotificationChannel, error) {
	row := q.db.QueryRow(ctx, createNotificationChannel, tenantID, name, channelType, config)
	var c NotificationChannel
	err := row.Scan(&c.ID, &c.TenantID, &c.Name, &c.ChannelType, &c.Config, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

const getNotificationChannel = `
SELECT id, tenant_id, name, channel_type, config, is_active, created_at, updated_at
FROM notification_channels WHERE id = $1 AND tenant_id = $2
`

func (q *Queries) GetNotificationChannel(ctx context.Context, id, tenantID uuid.UUID) (NotificationChannel, error) {
	row := q.db.QueryRow(ctx, getNotificationChannel, id, tenantID)
	var c NotificationChannel
	err := row.Scan(&c.ID, &c.TenantID, &c.Name, &c.ChannelType, &c.Config, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

const listNotificationChannels = `
SELECT id, tenant_id, name, channel_type, config, is_active, created_at, updated_at
FROM notification_channels WHERE tenant_id = $1 ORDER BY created_at DESC
`

func (q *Queries) ListNotificationChannels(ctx context.Context, tenantID uuid.UUID) ([]NotificationChannel, error) {
	rows, err := q.db.Query(ctx, listNotificationChannels, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []NotificationChannel
	for rows.Next() {
		var c NotificationChannel
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Name, &c.ChannelType, &c.Config, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	if items == nil {
		items = []NotificationChannel{}
	}
	return items, rows.Err()
}

const listActiveChannelsByTenant = `
SELECT id, tenant_id, name, channel_type, config, is_active, created_at, updated_at
FROM notification_channels WHERE tenant_id = $1 AND is_active = true
`

func (q *Queries) ListActiveChannelsByTenant(ctx context.Context, tenantID uuid.UUID) ([]NotificationChannel, error) {
	rows, err := q.db.Query(ctx, listActiveChannelsByTenant, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []NotificationChannel
	for rows.Next() {
		var c NotificationChannel
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Name, &c.ChannelType, &c.Config, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	if items == nil {
		items = []NotificationChannel{}
	}
	return items, rows.Err()
}

const deleteNotificationChannel = `DELETE FROM notification_channels WHERE id = $1 AND tenant_id = $2`

func (q *Queries) DeleteNotificationChannel(ctx context.Context, id, tenantID uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteNotificationChannel, id, tenantID)
	return err
}
