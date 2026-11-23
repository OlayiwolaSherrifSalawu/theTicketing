CREATE TABLE IF NOT EXISTS tickets(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quantity INT NOT NULL CHECK (quantity BETWEEN 1 AND 10),
    total_cents INT NOT NULL CHECK (total_cents >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'confirmed',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS tickets_user_id_idx ON tickets (user_id);
CREATE INDEX IF NOT EXISTS tickets_event_id_idx ON tickets (event_id);
