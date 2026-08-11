ALTER TABLE clicks
    DROP CONSTRAINT IF EXISTS clicks_code_fkey;

DROP INDEX IF EXISTS idx_urls_user_id_created_at;
DROP INDEX IF EXISTS idx_clicks_code_clicked_at;
