CREATE TABLE IF NOT EXISTS artist(
    artist_id PRIMARY KEY DEFAULT gen_random_uuid(), 
    name VARCHAR(255), 
    formation_date TIMESTAMP, 
    image VARCHAR(255),
    members TEXT[],
);