-- name: CreateNotification :one
INSERT INTO notifications (user_id, channel, recipient, subject, body, source)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateNotificationSuccess :exec
UPDATE notifications
SET 
  status = 'SENT', 
  sent_at = NOW(), 
  error_message = NULL
WHERE id = $1;

-- name: UpdateNotificationFailure :exec
UPDATE notifications
SET 
  status = 'FAILED', 
  error_message = $2, 
  retry_count = retry_count + 1
WHERE id = $1;

-- name: GetNotificationByID :one
SELECT * FROM notifications WHERE id = $1;

-- name: ListNotificationsByUserID :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC;
