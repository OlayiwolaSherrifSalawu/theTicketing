CREATE TABLE IF NOT EXISTS events(
    id uuid PRIMARY KEY gen_random_uuid(),
    location VARCHAR(255) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE,
    available_tickets INT NOT NULL,
    tickets_types VARCHAR(255) NOT NULL,
    event_name VARCHAR(255) NOT NULL,
    total_capacity INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);