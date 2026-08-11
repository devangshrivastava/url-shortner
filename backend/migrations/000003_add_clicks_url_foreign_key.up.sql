DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'clicks_code_fkey'
          AND conrelid = 'clicks'::regclass
    ) THEN
        ALTER TABLE clicks
            ADD CONSTRAINT clicks_code_fkey
            FOREIGN KEY (code) REFERENCES urls(code) ON DELETE CASCADE;
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_urls_user_id_created_at
    ON urls (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_clicks_code_clicked_at
    ON clicks (code, clicked_at DESC);
