CREATE TABLE IF NOT EXISTS artist(
    artist_id uuid PRIMARY KEY DEFAULT gen_random_uuid(), 
    name VARCHAR(255), 
    formation_date TEXT, 
    image VARCHAR(255),
    members TEXT[],
    first_album TEXT,
    events_id uuid REFERENCES events(id) ON DELETE CASCADE
);