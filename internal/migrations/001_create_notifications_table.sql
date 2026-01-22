DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'notification_status') THEN
            CREATE TYPE notification_status AS ENUM ('pending', 'scheduled', 'sent', 'failed', 'cancelled');
        END IF;
    END
$$;

CREATE TABLE IF NOT EXISTS notifications
(
    id          UUID PRIMARY KEY,
    message     TEXT                     NOT NULL,
    send_at     TIMESTAMP WITH TIME ZONE NOT NULL,
    status      notification_status      NOT NULL,
    channel     VARCHAR(20)              NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    retry_count INTEGER                  NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications (status);
CREATE INDEX IF NOT EXISTS idx_notifications_send_at ON notifications (send_at);
CREATE INDEX IF NOT EXISTS idx_notifications_status_send_at ON notifications (status, send_at);

COMMENT ON TABLE notifications IS 'Stores delayed notification requests';
COMMENT ON COLUMN notifications.id IS 'Unique identifier for the notification';
COMMENT ON COLUMN notifications.message IS 'Notification message content';
COMMENT ON COLUMN notifications.send_at IS 'Scheduled time for sending the notification';
COMMENT ON COLUMN notifications.status IS 'Current status of the notification (pending, scheduled, sent, failed, cancelled)';
COMMENT ON COLUMN notifications.channel IS 'Channel through which the notification will be sent (email, telegram, sms)';
COMMENT ON COLUMN notifications.created_at IS 'Timestamp when the notification was created';
COMMENT ON COLUMN notifications.updated_at IS 'Timestamp when the notification was last updated';
COMMENT ON COLUMN notifications.retry_count IS 'Number of retry attempts for failed notifications';;
