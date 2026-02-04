ALTER TABLE notifications ADD COLUMN email TEXT;
ALTER TABLE notifications ADD COLUMN telegram_id TEXT;

COMMENT ON COLUMN notifications.email IS 'Email address for email channel';
COMMENT ON COLUMN notifications.telegram_id IS 'Telegram user or chat ID for telegram channel';
