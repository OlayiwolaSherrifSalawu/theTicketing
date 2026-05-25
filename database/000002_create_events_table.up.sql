CREATE TABLE IF NOT EXISTS events(
    id uuid PRIMARY KEY,
    location VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    availableTicketss INT,
    ticketsTypes VARCHAR(255),
    eventName VARCHAR(255),
    totalCapacity INT
)