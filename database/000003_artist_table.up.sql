CREATE TABLE IF NOT EXISTS artist(
    artist_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255),
    formation_date TEXT,
    image VARCHAR(255),
    members TEXT [],
    first_album TEXT,
    external_id INT
);
ALTER TABLE events
ADD COLUMN artist_id uuid REFERENCES artist(artist_id) ON DELETE CASCADE