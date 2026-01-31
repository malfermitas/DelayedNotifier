DO
$$
    BEGIN
        IF NOT EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_name = 'notifications' AND column_name = 'email'
        ) THEN
            ALTER TABLE notifications ADD COLUMN email TEXT;
        END IF;
    END
$$;

DO
$$
    BEGIN
        IF NOT EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_name = 'notifications' AND column_name = 'telegram_id'
        ) THEN
            ALTER TABLE notifications ADD COLUMN telegram_id TEXT;
        END IF;
    END
$$;

COMMENT ON COLUMN notifications.email IS 'Email address for email channel';
COMMENT ON COLUMN notifications.telegram_id IS 'Telegram user or chat ID for telegram channel';
