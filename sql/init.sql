CREATE DATABASE scheduler;

CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    name TEXT,
    status TEXT,
    command TEXT,
    created_at,
);

-- CREATE INDEX idx_listings_link ON listings (link);