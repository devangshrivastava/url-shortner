ALTER TABLE urls
    ADD COLUMN IF NOT EXISTS user_id BIGINT,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'urls_user_id_fkey'
          AND conrelid = 'urls'::regclass
    ) THEN
        ALTER TABLE urls
            ADD CONSTRAINT urls_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES users(id);
    END IF;
END
$$;
