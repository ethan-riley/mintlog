-- name: CreateNotificationChannel :one
INSERT INTO notification_channels (tenant_id, name, channel_type, config)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetNotificationChannel :one
SELECT * FROM notification_channels WHERE id = $1 AND tenant_id = $2;

-- name: ListNotificationChannels :many
SELECT * FROM notification_channels WHERE tenant_id = $1 ORDER BY created_at DESC;

-- name: ListActiveChannelsByTenant :many
SELECT * FROM notification_channels WHERE tenant_id = $1 AND is_active = true;

-- name: DeleteNotificationChannel :exec
DELETE FROM notification_channels WHERE id = $1 AND tenant_id = $2;
