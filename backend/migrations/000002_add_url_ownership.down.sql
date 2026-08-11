ALTER TABLE urls
    DROP CONSTRAINT IF EXISTS urls_user_id_fkey,
    DROP COLUMN IF EXISTS user_id,
    DROP COLUMN IF EXISTS created_at;
