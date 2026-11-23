ALTER TABLE events
    ADD COLUMN organizer_id uuid REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN description TEXT NOT NULL DEFAULT '',
    ADD COLUMN category VARCHAR(50) NOT NULL DEFAULT 'Music',
    ADD COLUMN price_cents INT NOT NULL DEFAULT 0 CONSTRAINT events_price_non_negative CHECK (price_cents >= 0),
    ADD COLUMN image_url TEXT NOT NULL DEFAULT '',
    ADD CONSTRAINT events_available_within_capacity CHECK (available_tickets BETWEEN 0 AND total_capacity);

CREATE INDEX IF NOT EXISTS events_start_time_idx ON events (start_time);
CREATE INDEX IF NOT EXISTS events_organizer_id_idx ON events (organizer_id);
