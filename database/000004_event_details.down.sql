DROP INDEX IF EXISTS events_organizer_id_idx;
DROP INDEX IF EXISTS events_start_time_idx;
ALTER TABLE events
    DROP CONSTRAINT IF EXISTS events_available_within_capacity,
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS price_cents,
    DROP COLUMN IF EXISTS category,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS organizer_id;
