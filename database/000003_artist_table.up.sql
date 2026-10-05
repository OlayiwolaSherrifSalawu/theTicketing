CREATE TABLE IF NOT EXISTS artist(
    artist_id PRIMARY KEY DEFAULT gen_random_uuid() UUID, 
    name VARCHAR(255), 
    formation_date TEXT, 
    image VARCHAR(255),
    members TEXT[],
    first_album TEXT,
);