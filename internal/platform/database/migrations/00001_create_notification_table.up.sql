CREATE TYPE notification_channel AS ENUM (
  'EMAIL',
  'SMS',
  'PUSH',
  'TELEGRAM'
);

CREATE TYPE notification_status AS ENUM (
  'PENDING',
  'SENT',
  'DELIVERED',
  'FAILED',
  'READ',
  'BOUNCED'
);

CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    channel notification_channel NOT NULL,     -- SMS / EMAIL / PUSH
    recipient VARCHAR(128) NOT NULL,
    subject VARCHAR(255),
    body TEXT,
    status notification_status NOT NULL DEFAULT 'PENDING',
    source VARCHAR(64) NOT NULL,
    retry_count INT DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    sent_at TIMESTAMP
);
