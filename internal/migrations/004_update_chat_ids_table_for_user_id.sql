ALTER TABLE telegram_chats
    ADD COLUMN IF NOT EXISTS user_id TEXT;

UPDATE telegram_chats
SET user_id = cookie_id
WHERE user_id IS NULL;

ALTER TABLE telegram_chats
    ALTER COLUMN user_id SET NOT NULL;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE table_name = 'telegram_chats'
          AND constraint_type = 'PRIMARY KEY'
          AND constraint_name = 'telegram_chats_pkey'
    ) THEN
        ALTER TABLE telegram_chats DROP CONSTRAINT telegram_chats_pkey;
    END IF;
END $$;

ALTER TABLE telegram_chats
    ADD CONSTRAINT telegram_chats_pkey PRIMARY KEY (user_id);

DROP INDEX IF EXISTS telegram_chats_telegram_chat_id_idx;
CREATE INDEX IF NOT EXISTS telegram_chats_telegram_chat_id_idx ON telegram_chats (telegram_chat_id);
